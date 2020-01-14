package dao

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type UserExistsCallback func(exists bool, verifiedPassword bool, err error)
type CreateUserCallback func(err error)

const PEPPER string = "REGDQngkIasbXqT2@oWbcx42$ZwWF&@1d1or1k%p1F0YSfmAxHk5vxHJZp5D*Boh"

func UserExists(db *sql.DB, username string, password string, callback UserExistsCallback) {
	statement, err := db.Prepare("SELECT password FROM user WHERE login = ?")
	if err != nil {
		callback(false, false, err)
		return
	}
	defer statement.Close() // Close the statement when we leave main() / the program terminates
	var hashedPassword string
	err = statement.QueryRow(0).Scan(&hashedPassword)
	if err != nil || hashedPassword == "" {
		callback(false, false, nil)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password+PEPPER))
	if err != nil {
		callback(true, false, nil)
		return
	}
	callback(true, true, nil)
}

func CreateUser(db *sql.DB, username string, password string, email string, callback CreateUserCallback) {
	statement, err := db.Prepare("INSERT INTO user(login, password, email)  VALUES(?,?,?)")
	if err != nil {
		callback(err)
		return
	}
	defer statement.Close() // Close the statement when we leave main() / the program terminates

	hashedPassword, err := bcrypt.GenerateFromPassword(([]byte)(password+PEPPER), 10)
	if err != nil {
		callback(err)
		return
	}
	_, err = statement.Exec(username, hashedPassword, email)
	callback(err)
}
