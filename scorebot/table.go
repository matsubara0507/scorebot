package scorebot

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

type SQLInfo struct {
	SQLName                string
	CountUserSQL           func(userId string) string
	InsertUserSQL          func(userId string) string
	SelectUserSQL          func(userId string) string
	SelectAllUsersSQL      string
	UpdateUserSQL          func(userId string, value string) string
	SelectAllChallengesSQL string
}

func MakeUserTableImpl(databaseUrl string, challengesYaml string, sqlInfo SQLInfo) UserTable {
	return UserTable{
		SignUp: func(userId string) error {
			db, err := sql.Open(sqlInfo.SQLName, databaseUrl)
			defer db.Close()
			if err != nil {
				return err
			}
			cnt := 0
			query := sqlInfo.CountUserSQL(userId)
			err = db.QueryRow(query).Scan(&cnt)
			if err != nil {
				return err
			}
			if cnt != 0 {
				return nil
			}
			query = sqlInfo.InsertUserSQL(userId)
			_, err = db.Exec(query)
			return err
		},
		FindById: func(userId string) (*User, error) {
			db, err := sql.Open(sqlInfo.SQLName, databaseUrl)
			defer db.Close()
			if err != nil {
				return nil, err
			}
			challenges, err := ReadChallengesYaml(challengesYaml)
			if err != nil {
				return nil, err
			}
			user := MakeUser(userId, challenges)
			progress := make([]bool, len(user.Progress))
			var temp []interface{}
			temp = append(temp, &user.UserId)
			for i, _ := range progress {
				temp = append(temp, &progress[i])
			}
			query := sqlInfo.SelectUserSQL(userId)
			row := db.QueryRow(query)
			err = row.Scan(temp...)
			user = user.UpdateProgress(progress)
			return &user, err
		},
		FindAll: func() (*[]User, error) {
			db, err := sql.Open(sqlInfo.SQLName, databaseUrl)
			defer db.Close()
			if err != nil {
				return nil, err
			}
			rows, err := db.Query(sqlInfo.SelectAllUsersSQL)
			if err != nil {
				return nil, err
			}
			var users []User
			challenges, err := ReadChallengesYaml(challengesYaml)
			if err != nil {
				return nil, err
			}
			for rows.Next() {
				user := MakeUser("", challenges)
				progress := make([]bool, len(user.Progress))
				var temp []interface{}
				temp = append(temp, &user.UserId)
				for i, _ := range progress {
					temp = append(temp, &progress[i])
				}
				err := rows.Scan(temp...)
				if err != nil {
					return nil, err
				}
				users = append(users, user.UpdateProgress(progress))
			}
			return &users, nil
		},
		UpdateProgress: func(userId string, challengeId string, value bool) error {
			db, err := sql.Open(sqlInfo.SQLName, databaseUrl)
			defer db.Close()
			if err != nil {
				return err
			}
			query := sqlInfo.UpdateUserSQL(userId, fmt.Sprintf("%s = %v", value, userId))
			_, err = db.Exec(query)
			return err
		},
		ResetProgress: func(userId string) error {
			db, err := sql.Open(sqlInfo.SQLName, databaseUrl)
			defer db.Close()
			if err != nil {
				return err
			}
			challenges, err := ReadChallengesYaml(challengesYaml)
			if err != nil {
				return err
			}
			var values []string
			for _, k := range challenges.Keys() {
				values = append(values, k+" = false")
			}
			query := sqlInfo.UpdateUserSQL(userId, strings.Join(values, ", "))
			_, err = db.Exec(query)
			return err
		},
	}
}

func MakeChallengeTableImpl(databaseUrl string, sqlInfo SQLInfo) ChallengeTable {
	return ChallengeTable{
		FindAll: func(challenges *Challenges) error {
			db, err := sql.Open(sqlInfo.SQLName, databaseUrl)
			defer db.Close()
			if err != nil {
				return err
			}
			rows, err := db.Query(sqlInfo.SelectAllChallengesSQL)
			if err != nil {
				return err
			}
			for rows.Next() {
				var id, flag string
				err := rows.Scan(&id, &flag)
				if err != nil {
					return err
				}
				(*challenges)[id] = (*challenges)[id].SetFlag(flag)
			}
			return nil
		},
	}
}
