package services

import (
	"database/sql"

	"github.com/hoyle1974/grapevine/common"
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
	addr   common.Address
}

func (a AppCtx) Log(f string) zerolog.Logger {
	return a.log.With().Str("func", f).Logger()
}

func (a AppCtx) GetAddr() common.Address {
	return a.addr
}

func NewAppCtx(l zerolog.Logger, s *grpc.Server, db *sql.DB, addr common.Address) AppCtx {
	ctx := AppCtx{
		Server: s,
		log:    l,
		db:     db,
		addr:   addr,
	}

	ctx.log.Info().Msg("New Context created")

	return ctx
}
