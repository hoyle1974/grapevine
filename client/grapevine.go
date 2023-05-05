package main

import (
	"context"
	"fmt"
	"sync"

	pb "github.com/hoyle1974/grapevine/proto"
	"github.com/hoyle1974/grapevine/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

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
