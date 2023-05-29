package client

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	protoc "github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/hoyle1974/grapevine/proto"
	"github.com/hoyle1974/grapevine/services"
)

type Gossip interface {
	AddToGossip(rumor Rumor)
	GossipLoop(clientCache GrapevineClientCache)
	AddServer(contact services.ServerAddress)
}

type gossip struct {
	lock          sync.Mutex
	ctx           CallCtx
	self          services.ServerAddress
	rumors        Rumors
	mongers       GossipMongers
	knownSearches map[string]bool
}

func NewGossip(ctx CallCtx, self services.ServerAddress) Gossip {
	return &gossip{
		self:    self,
		ctx:     ctx.NewCtx("gossip"),
		mongers: NewGossipMongers(ctx, self),
		rumors:  NewRumors(ctx),
	}
}

func (g *gossip) AddServer(addr services.ServerAddress) {
	g.mongers.AddMonger(addr)
}

func (g *gossip) AddToGossip(rumor Rumor) {
	g.mongers.AddMonger(rumor.GetCreatorAddress())

	g.rumors.AddRumor(rumor)
}

func (g *gossip) getGossipRequest() *proto.GossipRequest {
	return g.rumors.GetProtobuf()
}

func (g *gossip) GossipLoop(clientCache GrapevineClientCache) {
	log := g.ctx.NewCtx("GossipLoop")

	for {
		time.Sleep(time.Second * 5)

		// Get a random address
		addr := g.mongers.GetRandomServerAddress()
		if addr == nil {
			// No one to gossip to
			log.Warn().Msg("No servers to gossip to")
			continue
		}

		g.lock.Lock()

		req := g.getGossipRequest()
		if req == nil {
			log.Warn().Msg("Nothing to gossip about, this is weird")
			continue
		}
		b, err := protoc.Marshal(req)
		if err != nil {
			log.Error().Err(err).Msg("Can't unmarshal")
			continue
		} else {
			log.Info().Msgf("Sending . . . ")

			contact := services.UserContact{
				Ip:   addr.GetIp(),
				Port: addr.GetPort(),
			}

			log.Info().Msgf("\tGossiping to %v:%v", contact.Ip, contact.Port)
			client := clientCache.GetClient(contact).GetClient()
			resp, err := client.Post(fmt.Sprintf("https://%s/gossip", contact.GetURL()), "grpc-message-type", bytes.NewReader(b))
			if err != nil {
				log.Error().Err(err).Msg("\tTried to post but got error")
				g.mongers.RemoveMonger(*addr)
				continue
			} else {
				log.Info().Msgf("\tResponse: %v", resp.StatusCode)
			}

			b, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Error().Err(err).Msg("\tError reading body of response")
			}

			gresp := proto.GossipResponse{}
			err = protoc.Unmarshal(b, &gresp)
			if err != nil {
				log.Error().Err(err).Msg("\tError unmarshaling response to our gossip")
			}

			for _, gossip := range gresp.Gossip {
				search := gossip.GetSearch()
				if search != nil {
					rumorId, err := uuid.Parse(search.SearchId)
					if err != nil {
						log.Warn().Err(err).Msgf("Couldn't parse: %v", search.SearchId)
						continue
					}

					rumor := NewSearchRumor(NewRumor(
						rumorId,
						gossip.EndOfLife.AsTime(),
						services.NewAccountId(search.Requestor.AccountId),
						services.NewServerAddress(net.ParseIP(search.Requestor.Address.IpAddress), search.Requestor.Address.Port),
					), search.Query)

					go g.AddToGossip(rumor)
				}
			}

		}

		g.lock.Unlock()
	}
}
