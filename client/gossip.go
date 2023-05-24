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
}

type gossip struct {
	lock         sync.Mutex
	ctx          CallCtx
	self         services.ServerAddress
	toGossip     []*proto.Gossip
	knownServers []services.ServerAddress
}

func NewGossip(ctx CallCtx, self services.ServerAddress) Gossip {
	return &gossip{self: self, ctx: ctx.NewCtx("gossip")}
}

func (g *gossip) AddServer(addr services.ServerAddress) {
	log := g.ctx.NewCtx("AddServer").Log()
	if net.IP.Equal(g.self.GetIp(), addr.GetIp()) && g.self.GetPort() == addr.GetPort() {
		return
	}
	g.lock.Lock()
	defer g.lock.Unlock()

	log.Printf("Adding server [%s:%d]", addr.GetIp(), addr.GetPort())
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
	return &proto.GossipRequest{Gossip: g.toGossip}
}

func (g *gossip) StartGossip(clientCache GrapevineClientCache) {
	log := g.ctx.NewCtx("AddServer")

	for {
		log.Printf("Gossip: sleep")
		time.Sleep(time.Second * 5)

		g.lock.Lock()
		log.Printf("Gossip: gossip: %v", g.toGossip)

		// Remove anything expired from the gossip chain
		for i := len(g.toGossip) - 1; i >= 0; i-- {
			if g.toGossip[i].EndOfLife.AsTime().After(time.Now()) {
				log.Printf("End of life found")
				g.toGossip = append(g.toGossip[:i], g.toGossip[i+1:]...)
			}
		}

		// Try to send the gossip chain to everyone we know about
		log.Printf("Getting request")
		req := g.getGossipRequest()
		b, err := protoc.Marshal(req)
		if err != nil {
			log.Error().Err(err).Msg("Can't unmarshal")
		} else {
			log.Printf("Sending (%v servers). . . ", len(g.knownServers))

			tmp := g.knownServers[:0]
			for _, addr := range g.knownServers {
				contact := services.UserContact{
					Ip:   addr.GetIp(),
					Port: addr.GetPort(),
				}

				log.Printf("Gossiping to %v", contact)
				client := clientCache.GetClient(contact).GetClient()
				resp, err := client.Post(fmt.Sprintf("https://%s/gossip", contact.GetURL()), "grpc-message-type", bytes.NewReader(b))
				if err != nil {
					log.Error().Err(err).Msg("Tried to post but got error")
				} else {
					log.Printf("Response: %v", resp)
					tmp = append(tmp, addr)
				}
			}
			g.knownServers = tmp
		}

		g.lock.Unlock()
	}
}
