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
	self         services.ServerAddress
	toGossip     []*proto.Gossip
	knownServers []services.ServerAddress
}

func NewGossip(self services.ServerAddress) Gossip {
	return &gossip{self: self}
}

func (g *gossip) AddServer(addr services.ServerAddress) {
	if net.IP.Equal(g.self.GetIp(), addr.GetIp()) && g.self.GetPort() == addr.GetPort() {
		return
	}
	g.lock.Lock()
	defer g.lock.Unlock()

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
	for {
		fmt.Println("Gossip: sleep")
		time.Sleep(time.Second * 5)

		g.lock.Lock()
		fmt.Printf("Gossip: gossip: %v\n", g.toGossip)

		// // Remove anything expired from the gossip chain
		// for i := len(g.toGossip) - 1; i >= 0; i-- {
		// 	if g.toGossip[i].EndOfLife.AsTime().After(time.Now()) {
		// 		fmt.Printf("End of life found")
		// 		g.toGossip = append(g.toGossip[:i], g.toGossip[i+1:]...)
		// 	}
		// }

		// Try to send the gossip chain to everyone we know about
		fmt.Printf("Getting request\n")
		req := g.getGossipRequest()
		b, err := protoc.Marshal(req)
		if err != nil {
			fmt.Println("Error " + err.Error())
		} else {
			fmt.Printf("Sending (%v servers). . . ", len(g.knownServers))
			for _, addr := range g.knownServers {
				contact := services.UserContact{
					Ip:   addr.GetIp(),
					Port: addr.GetPort(),
				}

				fmt.Printf("Gossiping to %v\n", contact)
				client := clientCache.GetClient(contact).GetClient()
				resp, err := client.Post(fmt.Sprintf("https://%s/gossip", contact.GetURL()), "grpc-message-type", bytes.NewReader(b))
				if err != nil {
					fmt.Printf("Tried to post but got error: %v\n", err.Error())
				} else {
					fmt.Printf("Response: %v\n", resp)
				}
			}
		}

		g.lock.Unlock()
	}
}
