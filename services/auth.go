package services

import (
	"net"

	"golang.org/x/crypto/bcrypt"
)

func Login(appCtx AppCtx, username string, password string, ip net.IP, port int32) (AccountId, error) {
	log := appCtx.Log("Login")
	log.Printf("Received: %v/%v", username, password)

	stmt := `select id, password_hash from "users" where "username"= $1`
	row := appCtx.db.QueryRow(stmt, username)

	if row.Err() != nil {
		return NilAccountId(), row.Err()
	}

	var id, hash string
	row.Scan(&id, &hash)

	var err error
	if err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return NilAccountId(), err
	}

	accountId := NewAccountId(id)

	err = UpdateUserContact(appCtx, accountId, ip, port)
	if err != nil {
		return NilAccountId(), err
	}

	return accountId, nil
}
