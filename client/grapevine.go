package client

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hoyle1974/grapevine/proto"
	"github.com/hoyle1974/grapevine/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Grapevine

type Grapevine interface {
	Start(ip net.IP) (int, error)
	Serve(s SharedData)
	JoinShare(s SharedData)
	LeaveShare(s SharedData)
	Invite(s SharedData, recipient services.UserContact, as string) bool
	Search(key string) SearchId

	CreateAccount(username string, password string) error
	Login(username string, password string, ip net.IP, port int) (services.AccountId, error)
}

type grapevine struct {
	lock        sync.Mutex
	ctx         CallCtx
	cb          ClientCallback
	listener    GrapevineListener
	clientCache GrapevineClientCache
	accountId   services.AccountId
	gossip      Gossip
}

func NewGrapevine(cb ClientCallback, ctx CallCtx) Grapevine {
	return &grapevine{cb: cb, ctx: ctx}
}

func (g *grapevine) Start(ip net.IP) (int, error) {
	ctx := g.ctx.NewCtx("Start")
	g.lock.Lock()
	defer g.lock.Unlock()

	ctx.Info().Msg("Starting grapevine . . ")

	// Start the server
	ctx.Info().Msg("Starting listener . . ")
	g.listener = NewGrapevineListener(ctx)
	port, err := g.listener.Listen(ip)
	if err != nil {
		return 0, err
	}

	// Create the client cache manager
	ctx.Info().Msg("Creating client cache manager . . ")
	g.clientCache = NewGrapevineClientCache()

	g.gossip = NewGossip(ctx, services.NewServerAddress(ip, int32(port)))
	go g.gossip.GossipLoop(g.clientCache)
	g.listener.SetGossip(g.gossip)

	g.gossip.AddServer(services.NewServerAddress(ip, 8911))

	return port, nil
}

// Services access
func (g *grapevine) CreateAccount(username string, password string) error {
	log := g.ctx.NewCtx("CreateAccount")

	log.Info().Msg("CreateAccount: " + *accountURL)
	conn, err := grpc.Dial(*accountURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := proto.NewAccountServiceClient(conn)

	_, err = client.CreateAccount(context.Background(), &proto.CreateAccountRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return err
	}
	return nil
}

func (g *grapevine) Login(username string, password string, ip net.IP, port int) (services.AccountId, error) {
	log := g.ctx.NewCtx("Login")

	log.Info().Msg("Login: " + *authURL)
	conn, err := grpc.Dial(*authURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return services.NilAccountId(), err
	}
	defer conn.Close()

	client := proto.NewAuthServiceClient(conn)

	resp, err := client.Auth(context.Background(), &proto.AuthRequest{
		Username:      username,
		Password:      password,
		ClientAddress: &proto.ClientAddress{IpAddress: ip.String(), Port: int32(port)},
	})
	if err != nil {
		return services.NilAccountId(), err
	}

	g.accountId = services.NewAccountId(resp.GetUserId())
	return g.accountId, nil
}

//------------------------------

func (g *grapevine) Serve(s SharedData) {
	// Make this shared data actually shareable
}

func (g *grapevine) JoinShare(s SharedData) {
	// Join a shared data
}

func (g *grapevine) LeaveShare(s SharedData) {
	// Leave a shared data
}

func (g *grapevine) Invite(s SharedData, recipient services.UserContact, as string) bool {
	// Invite someone to our shared data

	return false
}

// Initiating a search
func (g *grapevine) Search(query string) SearchId {
	log := g.ctx.NewCtx("Search")

	// Search using the gossip protocol
	log.Info().Msgf("Gossipping search for %v", query)

	rumor := NewSearchRumor(NewRumor(
		uuid.New(),
		time.Now().Add(time.Minute),
		g.accountId,
		services.NewServerAddress(g.listener.GetIp(), int32(g.listener.GetPort())),
	), query)

	g.gossip.AddToGossip(rumor)

	return SearchId(rumor.rumorId.String())

}
