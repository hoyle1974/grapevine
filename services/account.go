package services

import "github.com/hoyle1974/grapevine/models"

func CreateAccount(app AppCtx, username string, password string) (models.AccountId, error) {
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

	return models.NewAccountId(id), nil
}
