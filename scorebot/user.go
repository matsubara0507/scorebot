package scorebot

import (
	"sort"
)

type User struct {
	UserId   string
	Progress Progress
}

type Progress map[string]bool

func (user User) CalcScore(challenges Challenges) int {
	score := 0
	for k, v := range challenges {
		if user.Progress[k] {
			score = score + v.Point
		}
	}
	return score
}

func (user User) UpdateProgress(progress []bool) User {
	for i, k := range user.Progress.Keys() {
		if i < len(progress) {
			user.Progress[k] = progress[i]
		}
	}
	return user
}

func MakeUser(userId string, challenges Challenges) User {
	progress := map[string]bool{}
	for _, k := range challenges.Keys() {
		progress[k] = false
	}
	return User{
		UserId:   userId,
		Progress: progress,
	}
}

func (progress Progress) Keys() []string {
	var keys []string
	for k, _ := range progress {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

type UserTable struct {
	SignUp         func(userId string) error
	FindById       func(userId string) (*User, error)
	FindAll        func() (*[]User, error)
	UpdateProgress func(userId string, challengeId string, value bool) error
	ResetProgress  func(userId string) error
}
