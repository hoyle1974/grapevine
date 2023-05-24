package client

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"time"

	protoc "github.com/golang/protobuf/proto"
	"github.com/hoyle1974/grapevine/proto"
	"github.com/hoyle1974/grapevine/services"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Gossip interface {
	AddToGossip(proto interface{})
	StartGossip(clientCache GrapevineClientCache)
	AddServer(contact services.ServerAddress)
	RegisterSearchRequest(searchId SearchId) bool
}

type gossip struct {
	lock          sync.Mutex
	ctx           CallCtx
	self          services.ServerAddress
	toGossip      []*proto.Gossip
	knownServers  []services.ServerAddress
	knownSearches map[string]bool
}

func NewGossip(ctx CallCtx, self services.ServerAddress) Gossip {
	return &gossip{self: self, ctx: ctx.NewCtx("gossip")}
}

func (g *gossip) RegisterSearchRequest(searchId SearchId) bool {
	log := g.ctx.NewCtx("RegisterSearchRequest").Log()
	g.lock.Lock()
	defer g.lock.Unlock()

	_, ok := g.knownSearches[searchId.String()]
	if !ok {
		log.Info().Msgf("Register search id: %v", searchId)
		g.knownSearches[searchId.String()] = true
		return false
	}

	return true
}

func (g *gossip) AddServer(addr services.ServerAddress) {
	log := g.ctx.NewCtx("AddServer").Log()
	if net.IP.Equal(g.self.GetIp(), addr.GetIp()) && g.self.GetPort() == addr.GetPort() {
		log.Info().Msg("	\tskipping self")
		return
	}
	g.lock.Lock()
	defer g.lock.Unlock()

	// Do we know about this server?
	for _, s := range g.knownServers {
		if s.String() == addr.String() {
			return
		}
	}

	log.Info().Msgf("\tserver [%s:%d]", addr.GetIp(), addr.GetPort())
	g.knownServers = append(g.knownServers, addr)
}

func (g *gossip) AddToGossip(gsp interface{}) {
	g.lock.Lock()
	defer g.lock.Unlock()

	search := gsp.(*proto.Gossip_Search)

	g.toGossip = append(g.toGossip, &proto.Gossip{
		EndOfLife:   timestamppb.New(time.Now().Add(time.Hour)),
		GossipUnion: search,
	})
}

func (g *gossip) getGossipRequest() *proto.GossipRequest {
	if len(g.toGossip) == 0 {
		return nil
	}
	return &proto.GossipRequest{Gossip: g.toGossip}
}

func (g *gossip) clearGossip() {
	g.toGossip = []*proto.Gossip{}
}

func (g *gossip) StartGossip(clientCache GrapevineClientCache) {
	log := g.ctx.NewCtx("StartGossip")

	for {
		time.Sleep(time.Second * 5)

		g.lock.Lock()

		// Remove anything expired from the gossip chain
		// for i := len(g.toGossip) - 1; i >= 0; i-- {
		// 	if g.toGossip[i].EndOfLife.AsTime().After(time.Now()) {
		// 		log.Printf("End of life found: %v", g.toGossip[i])
		// 		g.toGossip = append(g.toGossip[:i], g.toGossip[i+1:]...)
		// 	}
		// }

		// Try to send the gossip chain to everyone we know about
		req := g.getGossipRequest()
		if req == nil {
			continue
		}
		b, err := protoc.Marshal(req)
		success := false
		if err != nil {
			log.Error().Err(err).Msg("Can't unmarshal")
		} else {
			log.Info().Msgf("Sending (%v servers). . . ", len(g.knownServers))

			tmp := g.knownServers[:0]
			for _, addr := range g.knownServers {
				contact := services.UserContact{
					Ip:   addr.GetIp(),
					Port: addr.GetPort(),
				}

				log.Info().Msgf("\tGossiping to %v:%v", contact.Ip, contact.Port)
				client := clientCache.GetClient(contact).GetClient()
				resp, err := client.Post(fmt.Sprintf("https://%s/gossip", contact.GetURL()), "grpc-message-type", bytes.NewReader(b))
				if err != nil {
					log.Error().Err(err).Msg("\tTried to post but got error")
				} else {
					log.Info().Msgf("\tResponse: %v", resp.StatusCode)
					tmp = append(tmp, addr)
					success = true
				}
			}
			g.knownServers = tmp
		}

		if success {
			g.clearGossip()
		}

		g.lock.Unlock()
	}
}
