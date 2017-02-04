package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/matsubara0507/scorebot/scorebot"
)

func main() {
	bot, err := linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
	if err != nil {
		log.Fatal(err)
	}
	databaseUrl := os.Getenv("DATABASE_URL")
	challengesYaml := "challenges.yaml"

	userTable := scorebot.MakeUserTableImplPostgreSQL(databaseUrl, challengesYaml)
	challengeTable := scorebot.MakeChallengeTableImplPostgreSQL(databaseUrl)

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
		challenges, err := scorebot.ReadChallengesYaml(challengesYaml)
		if err != nil {
			log.Print(err)
			return
		}
		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					texts := strings.Split(message.Text, " ")
					switch texts[0] {
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
			if event.Type == linebot.EventTypeFollow {
				err := userTable.SignUp(event.Source.UserID)
				if err == nil {
					bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("登録しました\n:navbar と入力してください")).Do()
				} else {
					log.Print(err)
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
