package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/earlgray283/kyopro_progress_reporter/util"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

var port string
var api *slack.Client
var members *[]Member
var channelID = "G01FQK55DPA"

func main() {
	var err error

	/*
		path, err := filepath.Abs(".env")
		if err != nil {
			log.Fatal(err)
		}
		if err := godotenv.Load(path); err != nil {
			log.Fatal(err)
		}
	*/

	api = slack.New(os.Getenv("SLACK_BOT_TOKEN"))
	if err := util.DownloadFile("members.json"); err != nil {
		log.Fatal(err)
	}
	members, err = NewMemberFromJSON()
	if err != nil {
		log.Fatal(err)
	}
	if err := ConvertIdToName(); err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := reportSubmissions(); err != nil {
			if _, _, err := api.PostMessage(channelID, slack.MsgOptionText(fmt.Sprintf("エラーが起きたっピ！朗読するっピ！\n%s", err.Error()), false)); err != nil {
				log.Println(err)
			}
			log.Fatal(err)
		}
	}()

	http.HandleFunc("/slack/events", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		switch eventsAPIEvent.Type {
		case slackevents.URLVerification:
			var res *slackevents.ChallengeResponse
			if err := json.Unmarshal(body, &res); err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "text/plain")
			if _, err := w.Write([]byte(res.Challenge)); err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

		case slackevents.CallbackEvent:
			innerEvent := eventsAPIEvent.InnerEvent

			switch event := innerEvent.Data.(type) {
			case *slackevents.AppMentionEvent:
				message := strings.Split(event.Text, " ")
				if len(message) < 2 {
					return
				}

				command := message[1]
				switch command {
				case "set":
					if _, _, err := api.PostMessage(event.Channel, slack.MsgOptionText("ここに json で設定を書き込む処理を入れるっぴ！", false)); err != nil {
						log.Println(err)
						w.WriteHeader(http.StatusInternalServerError)
					}

				default:
					if _, _, err := api.PostMessage(event.Channel, slack.MsgOptionText("何を言っているのかがわかんないっぴ！ごめんっぴ！", false)); err != nil {
						log.Println(err)
						w.WriteHeader(http.StatusInternalServerError)
					}
				}

				if _, _, err := api.PostMessage(event.Channel, slack.MsgOptionText("まだ週1報告以外の機能はできてないっぴ！ごめんなさいっぴ！", false)); err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
				}
			}
		}
	})

	log.Println("[INFO] Server listening...")
	port = os.Getenv("PORT")
	if err := http.ListenAndServe("0.0.0.0:"+port, nil); err != nil {
		log.Fatal(err)
	}

}
