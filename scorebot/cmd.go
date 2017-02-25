package scorebot

import (
	"fmt"
	"sort"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/mattn/sorter"
	"gopkg.in/kyokomi/emoji.v1"
)

type CmdType string

const (
	ScoreBotCmdChallenges CmdType = ":challenges"
	ScoreBotCmdHelp       CmdType = ":help"
	ScoreBotCmdNavBar     CmdType = ":navbar"
	ScoreBotCmdRanking    CmdType = ":ranking"
	ScoreBotCmdReset      CmdType = ":reset"
	ScoreBotCmdRule       CmdType = ":rule"
	ScoreBotCmdScore      CmdType = ":score"
)

func MakeChallengesMessage(challenges Challenges) []*linebot.CarouselTemplate {
	var columns []*linebot.CarouselColumn
	for _, k := range challenges.Keys() {
		action := linebot.NewMessageTemplateAction("More", ":challenges "+k)
		title := k + ": " + challenges[k].Title
		columns = append(columns, linebot.NewCarouselColumn("", "", title, action))
	}

	retVal := make([]*linebot.CarouselTemplate, len(columns)/5+1)
	for i, _ := range retVal {
		if i == len(columns)/5 {
			retVal[i] = linebot.NewCarouselTemplate(columns[i*5:]...)
		} else {
			retVal[i] = linebot.NewCarouselTemplate(columns[i*5 : (i+1)*5]...)
		}
	}
	return retVal
}

func MakeChallengeMessage(challengeId string, challenges Challenges) []linebot.Message {
	var messages []linebot.Message
	challenge, exist := challenges[challengeId]
	if !exist {
		message := linebot.NewTextMessage(challengeId + " is undefined as challenge.")
		return append(messages, message)
	}
	format := `%s: %s
Point: %d

%s`
	text := fmt.Sprintf(format, challengeId, challenge.Title, challenge.Point, challenge.Detail)
	messages = append(messages, linebot.NewTextMessage(text))
	if len(challenge.Choices) != 0 {
		title := fmt.Sprintf("%s: %s", challengeId, challenge.Title)
		var choicesMessages []linebot.TemplateAction
		for _, text := range challenge.Choices {
			choicesMessages = append(choicesMessages, linebot.NewMessageTemplateAction(text, text))
		}
		template := linebot.NewButtonsTemplate("", title, "次から選べ", choicesMessages...)
		messages = append(messages, linebot.NewTemplateMessage("選択肢", template))
	}
	return messages
}

func MakeRankingMessage(users []User, challenges Challenges, bot *linebot.Client) (*linebot.TextMessage, error) {
	type UserScore struct {
		UserId string
		Score  int
	}
	var usersWithScore []UserScore
	for _, user := range users {
		userWithScore := UserScore{user.UserId, user.CalcScore(challenges)}
		usersWithScore = append(usersWithScore, userWithScore)
	}
	sort.Sort(sorter.NewWrapperWith(
		usersWithScore,
		func(i, j int) bool {
			return usersWithScore[i].Score > usersWithScore[j].Score
		},
	))
	var messages []string
	for i, user := range usersWithScore {
		profile, err := bot.GetProfile(user.UserId).Do()
		if err != nil {
			return nil, err
		}
		message := fmt.Sprintf("%d : %s, %dpt", i+1, profile.DisplayName, user.Score)
		messages = append(messages, message)
	}
	return linebot.NewTextMessage(strings.Join(messages, "\n")), nil
}

func MakeScoreMessage(user User, challenges Challenges) *linebot.TextMessage {
	score := fmt.Sprintf("score: %d", user.CalcScore(challenges))
	messages := []string{score}
	for _, k := range challenges.Keys() {
		messages = append(messages, fmt.Sprintf("%s: %s", k, Checkbox(user.Progress[k])))
	}
	return linebot.NewTextMessage(strings.Join(messages, "\n"))
}

func MakeResultMessage(result bool) *linebot.TextMessage {
	var message string
	if result {
		message = "SUCCESS!!"
	} else {
		message = "FAILURE..."
	}
	return linebot.NewTextMessage(message)
}

func MakeNavBarMessage() *linebot.ButtonsTemplate {
	return linebot.NewButtonsTemplate(
		"",
		"ナビゲーションバー",
		"コマンド以外を入力するとフラグとして認識されます\n:help で他のコマンドも見れます",
		linebot.NewMessageTemplateAction("My Score", ":score"),
		linebot.NewMessageTemplateAction("Challenges", ":challenges"),
		linebot.NewMessageTemplateAction("Ranking", ":ranking"),
		linebot.NewMessageTemplateAction("Rule", ":rule"),
	)
}

func MakeHelpMessage() *linebot.TextMessage {
	var usage = `コマンド以外を入力するとフラグとして認識されます

コマンド リスト
:challenges  # 問題一覧
:help        # コマンド一覧
:navbar      # ナビゲーションバー
:ranking     # ランキングを表示
:reset       # スコアをリセット
:rule        # ルールを表示
:score       # スコアを表示`
	return linebot.NewTextMessage(usage)
}

func Checkbox(b bool) string {
	var checkbox string
	if b {
		checkbox = ":white_check_mark:"
	} else {
		checkbox = ":white_large_square:"
	}
	return emoji.Sprint(checkbox)
}
