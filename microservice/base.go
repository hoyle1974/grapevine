package microservice

import (
	"flag"
	"net"
	"os"

	grpczerolog "github.com/grpc-ecosystem/go-grpc-middleware/providers/zerolog/v2"
	middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	pb "github.com/hoyle1974/grapevine/proto"
	"github.com/hoyle1974/grapevine/services"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func ErrToProto(err error) *pb.Error {
	return &pb.Error{Msg: err.Error()}
}

func Start(name string, register func(appCtx services.AppCtx)) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if *env == "dev" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	l := log.With().Str("service", name).Logger()

	l.Info().Msg("Connecting to database")
	db, err := DB()
	if err != nil {
		l.Fatal().AnErr("Error connecting to database", err).Msg("error")
		return
	}

	l.Info().Msg("Starting " + name + " microservice")

	flag.Parse()
	addr := services.NewServerAddress(net.ParseIP(*ip), int32(*port))
	lis, err := net.Listen("tcp", addr.String())
	if err != nil {
		log.Fatal().AnErr("error", err).Msg("failed to listen")
	}

	s := grpc.NewServer(
		middleware.WithUnaryServerChain(
			logging.UnaryServerInterceptor(grpczerolog.InterceptorLogger(l)),
		),
		middleware.WithStreamServerChain(
			logging.StreamServerInterceptor(grpczerolog.InterceptorLogger(l)),
		),
	)

	appCtx := services.NewAppCtx(l, s, db, addr)

	register(appCtx)
	reflection.Register(s)
	l.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatal().AnErr("error", err).Msg("failed to serve")
	}

}
