// Copyright 2016 LINE Corporation
//
// LINE Corporation licenses this file to you under the Apache License,
// version 2.0 (the "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at:
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/linebot"
)

func main() {
	bot, err := linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Setup HTTP Server for receiving requests from LINE platform
	http.HandleFunc("/callback", func(w http.ResponseWriter, req *http.Request) {
		events, err := bot.ParseRequest(req)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(500)
			}
			return
		}
		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					texts := strings.Split(message.Text, " ")
					switch scorebot.CmdType(texts[0]) {
					case scorebot.ScoreBotCmdChallenges:
						if len(texts) <= 1 {
							challengesMessages := scorebot.MakeChallengesMessage(challenges)
							templateMessage := linebot.NewTemplateMessage("Challenges", challengesMessages[0])
							_, err = bot.ReplyMessage(event.ReplyToken, templateMessage).Do()
							if err != nil {
								log.Print(err)
							}
						} else {
							challengeMessage := scorebot.MakeChallengeMessage(texts[1], challenges)
							_, err = bot.ReplyMessage(event.ReplyToken, challengeMessage...).Do()
							if err != nil {
								log.Print(err)
							}
						}
					case scorebot.ScoreBotCmdScore:
						user, err := userTable.FindById(event.Source.UserID)
						if err == nil {
							scoreMessage := scorebot.MakeScoreMessage(*user, challenges)
							_, err = bot.ReplyMessage(event.ReplyToken, scoreMessage).Do()
							if err != nil {
								log.Print(err)
							}
						} else {
							log.Print(err)
						}
					case scorebot.ScoreBotCmdRanking:
						users, err := userTable.FindAll()
						if err == nil {
							rankingMessage, err := scorebot.MakeRankingMessage(*users, challenges, bot)
							if err == nil {
								_, err = bot.ReplyMessage(event.ReplyToken, rankingMessage).Do()
								if err != nil {
									log.Print(err)
								}
							} else {
								log.Print(err)
							}
						} else {
							log.Print(err)
						}
					case scorebot.ScoreBotCmdReset:
						err := userTable.ResetProgress(event.Source.UserID)
						if err == nil {
							_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("リセットしました")).Do()
							if err != nil {
								log.Print(err)
							}
						} else {
							log.Print(err)
						}
					case scorebot.ScoreBotCmdRule:
						_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("楽しんで")).Do()
						if err != nil {
							log.Print(err)
						}
					case scorebot.ScoreBotCmdNavBar:
						navBarMessage := scorebot.MakeNavBarMessage()
						templateMessage := linebot.NewTemplateMessage("ナビゲーションバー", navBarMessage)
						_, err = bot.ReplyMessage(event.ReplyToken, templateMessage).Do()
						if err != nil {
							log.Print(err)
						}
					case scorebot.ScoreBotCmdHelp:
						helpMessage := scorebot.MakeHelpMessage()
						_, err = bot.ReplyMessage(event.ReplyToken, helpMessage).Do()
						if err != nil {
							log.Print(err)
						}
					default:
						result, err := scorebot.Submit(
							scorebot.EqSubmitCondition(),
							message.Text, event.Source.UserID,
							challenges, userTable, challengeTable)
						if err == nil {
							resultMessage := scorebot.MakeResultMessage(result)
							_, err = bot.ReplyMessage(event.ReplyToken, resultMessage).Do()
							if err != nil {
								log.Print(err)
							}
						} else {
							log.Print(err)
						}
					}
				case *linebot.LocationMessage:
					result, err := scorebot.Submit(
						scorebot.NearLocationSubmitCondition(0.001, 0.001),
						fmt.Sprintf("%v %v", message.Latitude, message.Longitude), event.Source.UserID,
						challenges, userTable, challengeTable)
					if err == nil {
						message := scorebot.MakeResultMessage(result)
						_, err = bot.ReplyMessage(event.ReplyToken, message).Do()
						if err != nil {
							log.Print(err)
						}
					} else {
						log.Print(err)
					}
				case *linebot.StickerMessage:
					result, err := scorebot.Submit(
						scorebot.EqSubmitCondition(),
						fmt.Sprintf("sticker-%v-%v", message.PackageID, message.StickerID), event.Source.UserID,
						challenges, userTable, challengeTable)
					if err == nil {
						message := scorebot.MakeResultMessage(result)
						_, err = bot.ReplyMessage(event.ReplyToken, message).Do()
						if err != nil {
							log.Print(err)
						}
					} else {
						log.Print(err)
					}
				}
			}
		}
	})
	// This is just sample code.
	// For actual use, you must support HTTPS by using `ListenAndServeTLS`, a reverse proxy or something else.
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		log.Fatal(err)
	}
}
