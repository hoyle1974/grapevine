package client

import (
	"context"
	"net"
	"sync"

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
	server      GrapevineServer
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
	ctx.Info().Msg("Starting server component . . ")
	g.server = NewServer(ctx)
	port, err := g.server.Start(ip)
	if err != nil {
		return 0, err
	}

	// Create the client cache manager
	ctx.Info().Msg("Creating client cache manager . . ")
	g.clientCache = NewGrapevineClientCache()

	g.gossip = NewGossip(ctx, services.NewServerAddress(ip, int32(port)))
	go g.gossip.StartGossip(g.clientCache)
	g.server.SetGossip(g.gossip)

	gossipIP, err := net.LookupIP(*gossipAddr)
	if err != nil {
		ctx.Error().Caller().Msg("Unknown host: " + *gossipAddr)
	} else {
		ctx.Info().Msgf("Gossip (%s) IP address: %v", *gossipAddr, gossipIP)
		if len(gossipIP) > 0 {
			g.gossip.AddServer(services.NewServerAddress(gossipIP[0], 8911))
		}
	}
	// g.gossip.AddServer(services.NewServerAddress(net.ParseIP("10.42.0.130"), 8911))
	// g.gossip.AddServer(services.NewServerAddress(net.ParseIP("10.42.0.140"), 8911))
	// g.gossip.AddServer(services.NewServerAddress(net.ParseIP("10.42.0.150"), 8911))
	// g.gossip.AddServer(services.NewServerAddress(net.ParseIP("127.0.0.1"), 8911))

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

	// Create a search id
	searchId := SearchId(uuid.New().String())

	search := &proto.Search{
		SearchId: searchId.String(),
		Query:    query,
		Requestor: &proto.Contact{
			AccountId: g.accountId.String(),
			Address: &proto.ClientAddress{
				IpAddress: g.server.GetIp().String(),
				Port:      int32(g.server.GetPort()),
			},
		},
	}
	gossip_search := &proto.Gossip_Search{Search: search}

	g.gossip.AddToGossip(gossip_search)

	return searchId

}
