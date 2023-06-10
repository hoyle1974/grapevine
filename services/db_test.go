package services

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/hoyle1974/grapevine/common"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

var db *sql.DB

func TestMain(m *testing.M) {
	fmt.Println("Testing: TestMain")

	// os.Exit skips defer calls
	// so we need to call another function
	code, err := run(m)
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(code)
}

func NewTestAppCtx() AppCtx {
	l := log.With().Bool("test", true).Logger()

	if db == nil {
		initDB()
	}

	return NewAppCtx(l, nil, db, common.NewAddress(net.ParseIP("127.0.0.1"), 8911))
}

func initDB() {
	fmt.Println("Testing: run")
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, "postgres", "GAVxFWGABz", "grapevine")
	var err error
	db, err = sql.Open("postgres", psqlconn)
	if err != nil {
		panic(fmt.Errorf("could not connect to database: %w", err))
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		panic(fmt.Errorf("could not ping to database: %w", err))
	}
	fmt.Println("DB connected and pinged")

	schema, err := os.ReadFile("../schema.sql")
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(string(schema))
	if err != nil {
		panic(err)
	}
}

func run(m *testing.M) (int, error) {

	return m.Run(), nil
}
