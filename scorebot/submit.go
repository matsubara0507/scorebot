package scorebot

import (
	"math"
	"strconv"
	"strings"
)

func Submit(cond func(submit string, flag string) bool, flag string, userId string, challenges Challenges, userTable UserTable, challengeTable ChallengeTable) (bool, error) {
	err := challengeTable.FindAll(&challenges)
	if err != nil {
		return false, err
	}
	user, err := userTable.FindById(userId)
	if err != nil {
		return false, err
	}
	result, cid := challenges.Submit(cond, flag)
	if result {
		err = userTable.UpdateProgress(user.UserId, cid, result)
		if err != nil {
			return result, err
		}
	}
	return result, nil
}

func EqSubmitCondition() func(submit string, flag string) bool {
	return func(submit string, flag string) bool {
		return submit == flag
	}
}

func NearLocationSubmitCondition(x, y float64) func(submit string, flag string) bool {
	return func(submit string, flag string) bool {
		locF := strings.Split(flag, " ")
		if len(locF) != 2 {
			return false
		}
		locS := strings.Split(submit, " ")
		latitudeS, err1 := strconv.ParseFloat(locS[0], 64)
		longitudeS, err2 := strconv.ParseFloat(locS[1], 64)
		latitudeF, err3 := strconv.ParseFloat(locF[0], 64)
		longitudeF, err4 := strconv.ParseFloat(locF[1], 64)
		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			return false
		}
		return math.Abs(latitudeS-latitudeF) < x && math.Abs(longitudeS-longitudeF) < y
	}
}
