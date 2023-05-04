package services

import (
	"database/sql"

	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return string(bytes), err
}

type AppCtx struct {
	Server *grpc.Server
	log    zerolog.Logger
	db     *sql.DB
}

func (a AppCtx) Log(f string) zerolog.Logger {
	return a.log.With().Str("func", f).Logger()
}

func NewAppCtx(l zerolog.Logger, s *grpc.Server, db *sql.DB) AppCtx {
	ctx := AppCtx{
		Server: s,
		log:    l,
		db:     db,
	}

	ctx.log.Info().Msg("New Context created")

	return ctx
}
