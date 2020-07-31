package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/ChimeraCoder/anaconda"
)

const (
	firstPage int64 = -1
	countSize       = 200
)

type TwitterAPI struct {
	AccessToken       string `json:"accessToken"`
	AccessTokenSecret string `json:"accessTokenSecret"`
	ConsumerKey       string `json:"consumerKey"`
	ConsumerSecret    string `json:"consumerSecret"`
}

type SearchConditions struct {
	TargetScreenNames []string `json:"targetScreenNames"`
	TargetScreenName  string
	ExceptFollowing   bool   `json:"exceptFollowing"`
	ExceptFollowers   bool   `json:"exceptFollowers"`
	RunMode           string `json:"runMode"`
	myFollowers
}

type myFollowers []string

func init() {
	createCursorDir()
}

func main() {
	server := http.Server{Addr: ":8080"}
	http.Handle("/", http.FileServer(http.Dir("assets")))
	http.HandleFunc("/process/", process)
	server.ListenAndServe()
}

func process(w http.ResponseWriter, r *http.Request) {
	conds := extractConditions(r)
	if !conds.existsTargetScreenNames() {
		return
	}
	if conds.ExceptFollowers {
		conds.setList()
	}
	v := make(url.Values)
	v.Set("count", strconv.FormatInt(countSize, 10))

	for i := 0; i < len(conds.TargetScreenNames); i++ {
		conds.TargetScreenName = conds.TargetScreenNames[i]
		v.Set("screen_name", conds.TargetScreenName)
		ch := make(chan string, 3000)
		go conds.getScreenNames(v, ch)
		api := connectTwitterAPI()
		log.Printf("%s %s's followers", "Start blocking", conds.TargetScreenName)

		switch conds.RunMode {
		case "block":
			var blockCount int
			for {
				select {
				case screenName, ok := <-ch:
					if ok {
						api.BlockUser(screenName, nil)
						blockCount++
						if blockCount%500 == 0 {
							log.Printf("%d%s", blockCount, " users have been blocked")
							api = connectTwitterAPI()
						}
					} else {
						log.Printf("%s %d %s's %s", "Finally,", blockCount, conds.TargetScreenName, "users have been blocked")
						goto L
					}
				}
			}

		case "unblock":
			var unblockCount int
			for {
				select {
				case screenName, ok := <-ch:
					if ok {
						api.UnblockUser(screenName, nil)
						unblockCount++
						if unblockCount%500 == 0 {
							log.Printf("%d%s", unblockCount, " users have been unblocked")
							api = connectTwitterAPI()
						}
					} else {
						log.Printf("%s%d%s", "Finally, ", unblockCount, conds.TargetScreenName+"'s users have been unblocked")
						goto L
					}
				}
			}
		}
	L:
	}
	log.Println("Completed processing")
}
func (conds *SearchConditions) getScreenNames(v url.Values, ch chan string) {
	cursor := getSavedCursor(conds.TargetScreenName)
	for cursor != 0 {
		api := connectTwitterAPI()
		v.Set("cursor", strconv.FormatInt(cursor, 10))
		logRateLimitToFollowersList()
		c, err := api.GetFollowersList(v)
		if err != nil {
			log.Println(err)
		}

		for _, u := range c.Users {
			if conds.ExceptFollowing && u.Following {
				continue
			}
			if conds.ExceptFollowers && conds.containsTargetUser(u.ScreenName) {
				continue
			}
			ch <- u.ScreenName
		}

		cursor = c.Next_cursor
		if cursor == 0 {
			deleteCursorFile(conds.TargetScreenName)
		} else {
			saveCursor(conds.TargetScreenName, cursor)
		}
	}
	close(ch)
}

func getSavedCursor(targetScreenName string) int64 {
	if _, err := os.Stat("cursor/" + targetScreenName); err != nil {
		return firstPage
	}
	b, err := ioutil.ReadFile("cursor/" + targetScreenName)
	fmt.Println(b, err)
	if err != nil {
		log.Println(err)
	}
	s := string(b)
	if s == "" {
		return firstPage
	}
	cursor, err := strconv.Atoi(s)
	if err != nil {
		log.Println(err)
	}
	return int64(cursor)
}

func saveCursor(targetScreenName string, cursor int64) {
	s := strconv.Itoa(int(cursor))
	buf := make([]byte, 64)
	buf = []byte(s)
	err := ioutil.WriteFile("cursor/"+targetScreenName, buf, 0777)
	if err != nil {
		log.Println(err)
	}
}

func deleteCursorFile(targetScreenName string) {
	if _, err := os.Stat("cursor/" + targetScreenName); err != nil {
		return
	}
	if err := os.Remove("cursor/" + targetScreenName); err != nil {
		log.Println(err)
	}
}

func createCursorDir() {
	if _, err := os.Stat("cursor"); os.IsNotExist(err) {
		err := os.Mkdir("cursor", 0777)
		if err != nil {
			panic(err)
		}
	}
}

func logRateLimitToFollowersList() {
	api := connectTwitterAPI()
	ss := make([]string, 1)
	ss[0] = "followers"
	rateLimiteStatus, err := api.GetRateLimits(ss)
	if err != nil {
		log.Println(err)
	}
	br := rateLimiteStatus.Resources["followers"]["/followers/list"]
	log.Printf("%s %d %s\n", "Remaining", br.Remaining, "RateLimits of /followers/list")
	if br.Remaining == 0 {
		log.Println("Reached RateLimit")
		log.Println("Waiting for response...")
	}
}

func extractConditions(r *http.Request) (conds *SearchConditions) {
	body := r.Body
	defer body.Close()
	buf := new(bytes.Buffer)
	io.Copy(buf, body)
	json.Unmarshal(buf.Bytes(), &conds)
	return conds
}

func (conds *SearchConditions) existsTargetScreenNames() bool {
	if len(conds.TargetScreenNames) == 0 {
		log.Println("TargetScreenNames doesn't exist")
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
		logRateLimitToFollowersList()
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
