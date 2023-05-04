package services

func CreateAccount(app AppCtx, username string, password string) (AccountId, error) {
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

	return NewAccountId(id), nil
}
