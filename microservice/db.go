package microservice

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func DB() (*sql.DB, error) {
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", *dbhost, *dbport, *dbuser, *dbpassword, *dbname)

	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
