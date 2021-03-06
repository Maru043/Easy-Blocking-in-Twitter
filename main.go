package main

import (
	"bytes"
	"encoding/json"
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

	conds.createCursorFiles()
	for i := 0; i < len(conds.TargetScreenNames); i++ {
		conds.TargetScreenName = conds.TargetScreenNames[i]
		v.Set("screen_name", conds.TargetScreenName)
		ch := make(chan string, countSize)
		ch <- conds.TargetScreenName
		go conds.getScreenNames(v, ch)
		process2(conds.RunMode, conds.TargetScreenName, ch)
	}
	log.Println("Completed processing")
}

func process2(mode string, sn string, ch chan string) {
	log.Printf("%s %s's followers", mode, sn)
	api := connectTwitterAPI()
	proc := selectProc(api, mode)
	followersCount := findFollowersCount(sn)
	var cnt int
	for {
		select {
		case screenName, ok := <-ch:
			if ok {
				if _, err := proc(screenName, nil); err != nil {
					api = connectTwitterAPI()
					proc = selectProc(api, mode)
					proc(screenName, nil)
				}
				cnt++
				if cnt%200 == 0 {
					log.Printf("%d/%d %s's %s", cnt, followersCount, sn, "followers have been processed")
				}
			} else {
				log.Printf("%s %d %s's %s", "Finally,", cnt, sn, "followers have been processed")
				goto L
			}
		}
	}
L:
}

func selectProc(api *anaconda.TwitterApi, mode string) func(string, url.Values) (anaconda.User, error) {
	var proc func(string, url.Values) (anaconda.User, error)
	switch mode {
	case "block":
		proc = api.BlockUser
	case "unblock":
		proc = api.UnblockUser
	case "mute":
		proc = api.MuteUser
	case "unmute":
		proc = api.UnmuteUser
	}
	return proc
}

func (conds *SearchConditions) getScreenNames(v url.Values, ch chan string) {
	cursor := conds.getSavedCursor()
	for cursor != 0 {
		api := connectTwitterAPI()
		v.Set("cursor", strconv.FormatInt(cursor, 10))
		c, err := api.GetFollowersList(v)
		if err != nil {
			log.Println(err)
			break
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
			conds.deleteCursorFile()
		} else {
			conds.saveCursor(cursor)
		}
	}
	close(ch)
}

func (conds *SearchConditions) existsTargetScreenNames() bool {
	if len(conds.TargetScreenNames) == 0 {
		log.Println("TargetScreenNames doesn't exist")
		return false
	}
	return true
}

func (conds *SearchConditions) createCursorFiles() {
	for i := 0; i < len(conds.TargetScreenNames); i++ {
		targetScreenName := conds.TargetScreenNames[i]
		if _, err := os.Stat("cursor/" + targetScreenName); !os.IsNotExist(err) {
			continue
		}
		s := strconv.FormatInt(firstPage, 10)
		buf := []byte(s)
		if err := ioutil.WriteFile("cursor/"+targetScreenName, buf, 0666); err != nil {
			log.Println(err)
		}
	}
}

func (conds *SearchConditions) getSavedCursor() int64 {
	targetScreenName := conds.TargetScreenName
	if _, err := os.Stat("cursor/" + targetScreenName); err != nil {
		return firstPage
	}
	b, err := ioutil.ReadFile("cursor/" + targetScreenName)
	if err != nil {
		log.Println(err)
	}
	s := string(b)

	cursor, err := strconv.Atoi(s)
	if err != nil {
		log.Println(err)
	}
	return int64(cursor)
}

func (conds *SearchConditions) saveCursor(cursor int64) {
	targetScreenName := conds.TargetScreenName
	s := strconv.FormatInt(cursor, 10)
	buf := make([]byte, 64)
	buf = []byte(s)

	if err := ioutil.WriteFile("cursor/"+targetScreenName, buf, 0666); err != nil {
		log.Println(err)
	}
}

func (conds *SearchConditions) deleteCursorFile() {
	targetScreenName := conds.TargetScreenName
	if _, err := os.Stat("cursor/" + targetScreenName); err != nil {
		return
	}
	if err := os.Remove("cursor/" + targetScreenName); err != nil {
		log.Println(err)
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

func createCursorDir() {
	if _, err := os.Stat("cursor"); os.IsNotExist(err) {
		err := os.Mkdir("cursor", 0777)
		if err != nil {
			panic(err)
		}
	}
}

func (m *myFollowers) setList() {
	api := connectTwitterAPI()

	var cursor int64 = firstPage
	v := make(url.Values)
	v.Set("count", strconv.FormatInt(countSize, 10))

	log.Println("Setting my followers list...")
	for cursor != 0 {
		v.Set("cursor", strconv.FormatInt(cursor, 10))
		c, _ := api.GetFollowersList(v)
		for _, u := range c.Users {
			*m = append(*m, u.ScreenName)
		}
		cursor = c.Next_cursor
	}
	log.Println("Completed")
}

func (m myFollowers) containsTargetUser(s string) bool {
	for i := 0; i < len(m); i++ {
		if m[i] == s {
			return true
		}
	}
	return false
}

func findFollowersCount(screenName string) int {
	api := connectTwitterAPI()
	u, err := api.GetUsersShow(screenName, nil)
	if err != nil {
		log.Println(err)
		return 0
	}

	return u.FollowersCount
}

func connectTwitterAPI() *anaconda.TwitterApi {
	bytes, err := ioutil.ReadFile("./key.json")
	if err != nil {
		log.Fatal(err)
	}
	var twitterAPI TwitterAPI
	json.Unmarshal(bytes, &twitterAPI)
	api := anaconda.NewTwitterApiWithCredentials(twitterAPI.AccessToken, twitterAPI.AccessTokenSecret, twitterAPI.ConsumerKey, twitterAPI.ConsumerSecret)
	return api
}
