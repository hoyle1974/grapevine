package services

import (
	"database/sql"
	"fmt"
	"net"

	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return string(bytes), err
}

type ServerAddress struct {
	ip   net.IP
	port int32
}

func NewServerAddress(ip net.IP, port int32) ServerAddress {
	return ServerAddress{ip, port}
}

func (s ServerAddress) String() string {
	return fmt.Sprintf("%s:%d", s.ip.String(), s.port)
}

func (s ServerAddress) GetIp() net.IP {
	return s.ip
}

func (s ServerAddress) GetPort() int32 {
	return s.port
}

type AppCtx struct {
	Server *grpc.Server
	log    zerolog.Logger
	db     *sql.DB
	addr   ServerAddress
}

func (a AppCtx) Log(f string) zerolog.Logger {
	return a.log.With().Str("func", f).Logger()
}

func (a AppCtx) GetAddr() ServerAddress {
	return a.addr
}

func NewAppCtx(l zerolog.Logger, s *grpc.Server, db *sql.DB, addr ServerAddress) AppCtx {
	ctx := AppCtx{
		Server: s,
		log:    l,
		db:     db,
		addr:   addr,
	}

	ctx.log.Info().Msg("New Context created")

	return ctx
}
