package services

import (
	"net"

	"github.com/hoyle1974/grapevine/common"
	"golang.org/x/crypto/bcrypt"
)

func Auth(appCtx AppCtx, username string, password string, ip net.IP, port int32) (common.AccountId, error) {
	log := appCtx.Log("Login")
	log.Printf("Received: %v/%v", username, password)

	stmt := `select id, password_hash from "users" where "username"= $1`
	row := appCtx.db.QueryRow(stmt, username)

	if row.Err() != nil {
		return common.NilAccountId(), row.Err()
	}

	var id, hash string
	row.Scan(&id, &hash)

	var err error
	if err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return common.NilAccountId(), err
	}

	accountId := common.NewAccountId(id)

	err = UpdateUserContact(appCtx, accountId, ip, port)
	if err != nil {
		return common.NilAccountId(), err
	}

	return accountId, nil
}
