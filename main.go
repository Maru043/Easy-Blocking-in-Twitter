package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ChimeraCoder/anaconda"
)

const (
	firstPage = -1
	countSize = 200
)

type TwitterAPI struct {
	AccessToken       string `json:"accessToken"`
	AccessTokenSecret string `json:"accessTokenSecret"`
	ConsumerKey       string `json:"consumerKey"`
	ConsumerSecret    string `json:"consumerSecret"`
}

type SearchConditions struct {
	TargetScreenName string `json:"targetScreenName"`
	ExceptFollowing  bool   `json:"exceptFollowing"`
	ExceptFollowers  bool   `json:"exceptFollowers"`
	myFollowers
}

type myFollowers []string

func main() {
	server := http.Server{Addr: ":8080"}
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("assets"))))
	http.HandleFunc("/block/", blockUsers)
	http.HandleFunc("/unblock/", unblockUsers)
	server.ListenAndServe()
}

func blockUsers(w http.ResponseWriter, r *http.Request) {
	conds := parseRequest(r)
	if !conds.existsTargetScreenName() {
		return
	}
	if conds.ExceptFollowers {
		conds.setList()
	}
	v := make(url.Values)
	v.Set("screen_name", conds.TargetScreenName)
	v.Set("count", strconv.FormatInt(countSize, 10))
	screenNames := conds.getScreenNamesToBlock(v)
	api := connectTwitterAPI()
	for i := 0; i < len(screenNames); i++ {
		api.BlockUser(screenNames[i], nil)
	}
}

func unblockUsers(w http.ResponseWriter, r *http.Request) {
	conds := parseRequest(r)
	v := make(url.Values)
	v.Set("screen_name", conds.TargetScreenName)
	v.Set("count", strconv.FormatInt(countSize, 10))
	screenNames := conds.getScreenNamesToUnblock(v)
	api := connectTwitterAPI()
	for i := 0; i < len(screenNames); i++ {
		api.UnblockUser(screenNames[i], nil)
	}
}

func parseRequest(r *http.Request) (conds *SearchConditions) {
	body := r.Body
	defer body.Close()
	buf := new(bytes.Buffer)
	io.Copy(buf, body)
	json.Unmarshal(buf.Bytes(), &conds)
	return conds
}

func (conds *SearchConditions) existsTargetScreenName() bool {
	targetScreenName := conds.TargetScreenName
	if targetScreenName == "" {
		log.Println("TargetScreenName doesn't exist")
		return false
	}
	return true
}

func (m myFollowers) setList() {
	api := connectTwitterAPI()
	var cursor int64 = firstPage
	v := make(url.Values)
	v.Set("count", strconv.FormatInt(countSize, 10))
	for cursor != 0 {
		v.Set("cursor", strconv.FormatInt(cursor, 10))
		c, _ := api.GetFollowersList(v)
		for _, u := range c.Users {
			m = append(m, u.ScreenName)
		}
		cursor = c.Next_cursor
	}
	return
}

func (m myFollowers) containsTargetUser(s string) bool {
	for i := 0; i < len(m); i++ {
		if m[i] == s {
			return true
		}
	}
	return false
}

func (conds *SearchConditions) getScreenNamesToBlock(v url.Values) (screenNames []string) {
	api := connectTwitterAPI()
	var cursor int64 = firstPage
	for cursor != 0 {
		v.Set("cursor", strconv.FormatInt(cursor, 10))
		c, _ := api.GetFollowersList(v)
		for _, u := range c.Users {
			if conds.ExceptFollowing && u.Following {
				continue
			}
			if conds.ExceptFollowers && conds.containsTargetUser(u.ScreenName) {
				continue
			}
			screenNames = append(screenNames, u.ScreenName)
		}
		cursor = c.Next_cursor
	}
	return
}

func (conds *SearchConditions) getScreenNamesToUnblock(v url.Values) (screenNames []string) {
	api := connectTwitterAPI()
	var cursor int64 = firstPage
	for cursor != 0 {
		v.Set("cursor", strconv.FormatInt(cursor, 10))
		c, _ := api.GetFollowersList(v)
		for _, u := range c.Users {
			screenNames = append(screenNames, u.ScreenName)
		}
		cursor = c.Next_cursor
	}
	return
}

func connectTwitterAPI() *anaconda.TwitterApi {
	bytes, err := ioutil.ReadFile("./key.json")
	if err != nil {
		panic(err)
	}
	var twitterAPI TwitterAPI
	json.Unmarshal(bytes, &twitterAPI)
	api := anaconda.NewTwitterApiWithCredentials(twitterAPI.AccessToken, twitterAPI.AccessTokenSecret, twitterAPI.ConsumerKey, twitterAPI.ConsumerSecret)
	return api
}
