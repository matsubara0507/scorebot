package scorebot

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

func MakeUserTableImplPostgreSQL(databaseUrl string, challengesYaml string) UserTable {
	return UserTable{
		SignUp: func(userId string) error {
			db, err := sql.Open("postgres", databaseUrl)
			defer db.Close()
			if err != nil {
				return err
			}
			cnt := 0
			err = db.QueryRow("SELECT COUNT(*) FROM users WHERE userid = $1;", userId).Scan(&cnt)
			if err != nil {
				return err
			}
			if cnt != 0 {
				return nil
			}
			_, err = db.Exec("INSERT INTO users VALUES ($1);", userId)
			return err
		},
		FindById: func(userId string) (*User, error) {
			db, err := sql.Open("postgres", databaseUrl)
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
			row := db.QueryRow("SELECT * FROM users WHERE userid = $1;", userId)
			err = row.Scan(temp...)
			user = user.UpdateProgress(progress)
			return &user, err
		},
		FindAll: func() (*[]User, error) {
			db, err := sql.Open("postgres", databaseUrl)
			defer db.Close()
			if err != nil {
				return nil, err
			}
			rows, err := db.Query("SELECT * FROM users;")
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
			db, err := sql.Open("postgres", databaseUrl)
			defer db.Close()
			if err != nil {
				return err
			}
			_, err = db.Exec(fmt.Sprintf("UPDATE users SET %s = $1 WHERE userid = $2", challengeId), value, userId)
			return err
		},
		ResetProgress: func(userId string) error {
			db, err := sql.Open("postgres", databaseUrl)
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
			_, err = db.Exec(fmt.Sprintf("UPDATE users SET %s WHERE userid = $1", strings.Join(values, ", ")), userId)
			return err
		},
	}
}

func MakeChallengeTableImplPostgreSQL(databaseUrl string) ChallengeTable {
	return ChallengeTable{
		FindAll: func(challenges *Challenges) error {
			db, err := sql.Open("postgres", databaseUrl)
			defer db.Close()
			if err != nil {
				return err
			}
			rows, err := db.Query("SELECT * FROM challenges;")
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
