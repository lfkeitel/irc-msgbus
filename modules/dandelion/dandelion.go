package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/lfkeitel/yobot/ircbot"
	"github.com/lfkeitel/yobot/plugins"

	"github.com/lfkeitel/yobot/config"
	"github.com/lfkeitel/yobot/utils"
)

type dandelionConfig struct {
	URL, ApiKey string
	Channels    []string
}

var (
	dconf  *dandelionConfig
	lastID = 0
)

func init() {
	plugins.RegisterInit(dandelionInit)
}

func dandelionInit(conf *config.Config, bot *ircbot.Conn) {
	var dc dandelionConfig
	if err := utils.FillStruct(&dc, conf.Modules["dandelion"]); err != nil {
		panic(err)
	}
	dconf = &dc

	go startDandelionCheck()
}

type dandelionResp struct {
	Status      string
	Errorcode   int
	Module      string
	RequestTime string
	Data        map[string]dandelionLog
}

type dandelionLog struct {
	ID          int
	DateCreated string
	TimeCreated string
	Title       string
	Body        string
	UserID      int
	Category    string
	IsEdited    int
	Fullname    string
	CanEdit     bool
	// On metadata key only
	Limit       int
	LogSize     int
	Offset      int
	ResultCount int
}

func startDandelionCheck() {
	readAPI := dconf.URL + "/api/logs/read"
	params := make(url.Values)
	params.Set("apikey", dconf.ApiKey)
	params.Set("limit", "10")
	readAPI = readAPI + "?" + params.Encode()

	for {
		var decoder *json.Decoder
		var apiResp dandelionResp
		var logs []dandelionLog
		var irc *ircbot.Conn

		fmt.Println("Checking Dandelion")

		resp, err := http.Get(readAPI)
		if err != nil {
			fmt.Println(err)
			goto sleep
		}
		defer resp.Body.Close()

		decoder = json.NewDecoder(resp.Body)
		if err := decoder.Decode(&apiResp); err != nil {
			fmt.Println(err)
			goto sleep
		}

		// Bad API request
		if apiResp.Errorcode != 0 {
			fmt.Println(apiResp.Status)
			goto sleep
		}

		// No returned logs
		if apiResp.Data["metadata"].ResultCount == 0 {
			goto sleep
		}

		if apiResp.Data["0"].ID <= lastID {
			goto sleep
		}

		logs = make([]dandelionLog, 0, len(apiResp.Data)-1)

		for key, log := range apiResp.Data {
			if key == "metadata" {
				continue
			}

			if log.ID > lastID {
				logs = append(logs, log)
			}
		}

		irc = ircbot.GetBot()
		for _, log := range logs {
			for _, channel := range dconf.Channels {
				irc.Privmsgf(channel, "Dandelion - %s (%s) <%s/logs/%d>", log.Title, log.Fullname, dconf.URL, log.ID)
			}
		}

	sleep:
		time.Sleep(10 * time.Second)
	}
}

func main() {}