package scorebot

import (
	"fmt"

	_ "github.com/lib/pq"
)

func PostgreSQLInfo() SQLInfo {
	return SQLInfo{
		SQLName: "postgres",
		CountUserSQL: func(userId string) string {
			return fmt.Sprintf("SELECT COUNT(*) FROM users WHERE userid = '%s';", userId)
		},
		InsertUserSQL: func(userId string) string {
			return fmt.Sprintf("INSERT INTO users(userid) VALUES ('%s');", userId)
		},
		SelectUserSQL: func(userId string) string {
			return fmt.Sprintf("SELECT * FROM users WHERE userid = '%s';", userId)
		},
		SelectAllUsersSQL: "SELECT * FROM users;",
		UpdateUserSQL: func(userId string, value string) string {
			return fmt.Sprintf("UPDATE users SET %s WHERE userid = '%s'", value, userId)
		},
		SelectAllChallengesSQL: "SELECT * FROM challenges;",
	}
}

func MakeUserTableImplPostgreSQL(databaseUrl string, challengesYaml string) UserTable {
	return MakeUserTableImpl(databaseUrl, challengesYaml, PostgreSQLInfo())
}

func MakeChallengeTableImplPostgreSQL(databaseUrl string) ChallengeTable {
	return MakeChallengeTableImpl(databaseUrl, PostgreSQLInfo())
}
