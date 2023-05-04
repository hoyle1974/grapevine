package client

import (
	"sync"

	pb "github.com/hoyle1974/grapevine/proto"
	"github.com/hoyle1974/grapevine/services"
)

type Grapevine interface {
	OnGossipSearch(appCtx services.AppCtx, requstor services.UserContact, query string, orig *pb.Search)
}

type GrapevineClientCallback interface {
	Search(query string) bool
}

func NewGrapevine(cb GrapevineClientCallback) Grapevine {
	return &grapevine{cb: cb}
}

type grapevine struct {
	lock       sync.Mutex
	cb         GrapevineClientCallback
	gossipList []interface{}
}

func (g *grapevine) OnGossipSearch(appCtx services.AppCtx, requstor services.UserContact, query string, orig *pb.Search) {
	g.lock.Lock()
	defer g.lock.Unlock()

	//Process the query, if we have an answer, respond, otherwise add to outgoing gossip
	if g.cb.Search(query) {
		// Notify the requestor that we are a match for the query
		// TODO
	}

	// Add to gossip list
	g.gossipList = append(g.gossipList, orig)

}
