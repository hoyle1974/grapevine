package services

import "github.com/hoyle1974/grapevine/common"

func CreateAccount(app AppCtx, username string, password string) (common.AccountId, error) {
	log := app.Log("CreateAccount")
	log.Printf("Received: %v/****", username)

	hash, err := HashPassword(password)
	if err != nil {
		return "", err
	}

	stmt := `insert into "users"("id", "username","password_hash") values(gen_random_uuid(),$1, $2) returning id`
	row := app.db.QueryRow(stmt, username, hash)
	if row.Err() != nil {
		return "", err
	}

	var id string
	row.Scan(&id)

	return common.NewAccountId(id), nil
}
