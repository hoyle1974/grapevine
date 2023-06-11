package client

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hoyle1974/grapevine/common"
	"github.com/hoyle1974/grapevine/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Grapevine

type Grapevine interface {
	Start(ip net.IP) (int, error)
	Serve(s SharedData) SharedData
	JoinShare(s SharedData)
	LeaveShare(s SharedData)
	Invite(s SharedData, recipient common.Contact, as string) bool
	Search(key string) SearchId
	GetMe() common.Contact

	CreateAccount(username string, password string) error
	Login(username string, password string, ip net.IP, port int) (common.AccountId, error)
}

type grapevine struct {
	lock              sync.Mutex
	ctx               CallCtx
	cb                ClientCallback
	listener          GrapevineListener
	clientCache       GrapevineClientCache
	accountId         common.AccountId
	gossip            Gossip
	sharedDataManager SharedDataManager
}

func NewGrapevine(cb ClientCallback, ctx CallCtx) Grapevine {
	return &grapevine{cb: cb, ctx: ctx}
}

func (g *grapevine) GetMe() common.Contact {
	return g.listener.GetMe()
}

func (g *grapevine) Start(ip net.IP) (int, error) {
	ctx := g.ctx.NewCtx("Start")
	g.lock.Lock()
	defer g.lock.Unlock()

	ctx.Info().Msg("Starting grapevine . . ")

	// Start the server
	ctx.Info().Msg("Starting listener . . ")
	onSearchCB := func(searchId SearchId, query string) bool {
		return g.cb.OnSearch(searchId, query)
	}
	onSearchResultCB := func(searchId SearchId, response string, accountId common.AccountId, ip string, port int) {
		g.cb.OnSearchResult(searchId, response, common.NewContact(accountId, net.ParseIP(ip), port))
	}
	g.listener = NewGrapevineListener(ctx, onSearchCB, onSearchResultCB)
	port, err := g.listener.Listen(ip)
	if err != nil {
		return 0, err
	}

	// Create the client cache manager
	ctx.Info().Msg("Creating client cache manager . . ")
	g.clientCache = NewGrapevineClientCache()
	g.listener.SetClientCache(g.clientCache)

	g.sharedDataManager = NewSharedDataManager(ctx, g.listener, g.cb, g.clientCache)
	g.listener.SetSharedDataManager(g.sharedDataManager)

	g.gossip = NewGossip(ctx, common.NewAddress(ip, port))
	go g.gossip.GossipLoop(g.clientCache)
	g.listener.SetGossip(g.gossip)

	ctx.Info().Msgf("Adding server %v", ip)
	g.gossip.AddServer(common.NewAddress(ip, 8911))

	// addrs, err := net.LookupHost(*gossipAddr)
	// if err == nil {
	// 	if len(addrs) == 0 {
	// 		ctx.Error().Msgf("No addresses found for %s", *gossipAddr)
	// 	} else {
	// 		ip := net.ParseIP(addrs[0])
	// 		ctx.Info().Msgf("Adding server %v from host %v", ip, addrs[0])
	// 		g.gossip.AddServer(services.NewServerAddress(ip, 8911))
	// 	}
	// } else {
	// 	ctx.Error().Err(err).Msgf("There was an error looking up %s", *gossipAddr)
	// }

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

func (g *grapevine) Login(username string, password string, ip net.IP, port int) (common.AccountId, error) {
	log := g.ctx.NewCtx("Login")

	log.Info().Msg("Login: " + *authURL)
	conn, err := grpc.Dial(*authURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return common.NilAccountId(), err
	}
	defer conn.Close()

	client := proto.NewAuthServiceClient(conn)

	resp, err := client.Auth(context.Background(), &proto.AuthRequest{
		Username:      username,
		Password:      password,
		ClientAddress: &proto.ClientAddress{IpAddress: ip.String(), Port: int32(port)},
	})
	if err != nil {
		return common.NilAccountId(), err
	}

	g.accountId = common.NewAccountId(resp.GetUserId())
	g.listener.SetAccountId(g.accountId)
	return g.accountId, nil
}

//------------------------------

func (g *grapevine) Serve(s SharedData) SharedData {
	// Make this shared data actually shareable
	return g.sharedDataManager.Serve(s)
}

func (g *grapevine) JoinShare(s SharedData) {
	// Join a shared data
	g.sharedDataManager.JoinShare(s)
}

func (g *grapevine) LeaveShare(s SharedData) {
	// Leave a shared data
	g.sharedDataManager.LeaveShare(s)
}

func (g *grapevine) Invite(s SharedData, recipient common.Contact, as string) bool {
	// Invite someone to our shared data
	return g.sharedDataManager.Invite(s, recipient, as)
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
		common.NewAddress(g.listener.GetIp(), g.listener.GetPort()),
	), query)

	g.gossip.AddToGossip(rumor)

	return SearchId(rumor.rumorId.String())

}
