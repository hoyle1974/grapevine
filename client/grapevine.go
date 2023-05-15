package client

import (
	"context"
	"net"
	"sync"

	"github.com/google/uuid"
	"github.com/hoyle1974/grapevine/proto"
	"github.com/hoyle1974/grapevine/services"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const ACCOUNT_URL = "localhost:8080"
const AUTH_URL = "localhost:8081"

/*
type Grapevine interface {
	OnGossipSearch(appCtx services.AppCtx, requstor services.UserContact, query string, orig *pb.Search)
}

type GrapevineClientCallback interface {
	OnSearchQuery(query string) bool
}

func NewGrapevine(cb GrapevineClientCallback, accountId services.AccountId, addr services.ServerAddress) Grapevine {
	myContact := &pb.Contact{
		AccountId: accountId.String(),
		Address: &pb.ClientAddress{
			IpAddress: addr.GetIp().String(),
			Port:      addr.GetPort(),
		},
	}
	return &grapevine{cb: cb, myContact: myContact}
}

type grapevine struct {
	lock       sync.Mutex
	cb         GrapevineClientCallback
	gossipList []interface{}
	myContact  *pb.Contact
}

func (g *grapevine) getClientFor(contact services.UserContact) (pb.GrapevineServiceClient, error) {

	url := fmt.Sprintf("%v:%d", contact.Ip, contact.Port)
	conn, err := grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return pb.NewGrapevineServiceClient(conn), nil
}

func (g *grapevine) OnGossipSearch(appCtx services.AppCtx, requestor services.UserContact, query string, orig *pb.Search) {
	g.lock.Lock()
	defer g.lock.Unlock()

	//Process the query, if we have an answer, respond, otherwise add to outgoing gossip
	if g.cb.OnSearchQuery(query) {
		// Notify the requestor that we are a match for the query
		client, err := g.getClientFor(requestor)
		if err != nil {
			client.SearchResult(context.Background(), &pb.SearchResultRequest{
				SearchId:  orig.SearchId,
				Responder: g.myContact,
				Response:  "",
			})
		}
	}

	// Add to gossip list
	g.gossipList = append(g.gossipList, orig)

}
*/

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
	cb          ClientCallback
	server      GrapevineServer
	clientCache GrapevineClientCache
	accountId   services.AccountId
	gossip      Gossip
}

func NewGrapevine(cb ClientCallback) Grapevine {
	return &grapevine{cb: cb}
}

func (g *grapevine) Start(ip net.IP) (int, error) {
	g.lock.Lock()
	defer g.lock.Unlock()

	log.Info().Msg("Starting grapevine . . ")

	// Start the server
	log.Info().Msg("Starting server component . . ")
	g.server = NewServer()
	port, err := g.server.Start(ip)
	if err != nil {
		return 0, err
	}

	// Create the client cache manager
	log.Info().Msg("Creating client cache manager . . ")
	g.clientCache = NewGrapevineClientCache()

	g.gossip = NewGossip(services.NewServerAddress(ip, int32(port)))
	go g.gossip.StartGossip(g.clientCache)

	g.gossip.AddServer(services.NewServerAddress(ip, 8911))

	return port, nil
}

// Services access
func (g *grapevine) CreateAccount(username string, password string) error {
	conn, err := grpc.Dial(ACCOUNT_URL, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
	conn, err := grpc.Dial(AUTH_URL, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
