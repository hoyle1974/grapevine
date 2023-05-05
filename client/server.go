package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

type GrapevineServer interface {
	Start() error
}

type grapevineServer struct {
}

func NewServer() GrapevineServer {
	return &grapevineServer{}
}

/*
service GrapevineService {
  rpc Gossip (GossipRequest) returns (GossipResponse);

  rpc SearchResult (SearchResultRequest) returns (SearchResultResponse);

  rpc SharedInvitation (SharedInvitationRequest) returns (SharedInvitationResponse);
  rpc ChangeDataOwner (ChangeDataOwnerRequest) returns (ChangeDataOwnerResponse);
  rpc ChangeData (ChangeDataRequest) returns (ChangeDataResponse);
  rpc LeaveSharedData (LeaveSharedDataRequest) returns (LeaveSharedDataResponse);
}

*/

func (g *grapevineServer) gossip(writer http.ResponseWriter, req *http.Request) {
	/*
		mux.HandleFunc("/distribute", func(writer http.ResponseWriter, req *http.Request) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				fmt.Printf("error reading body while handling /distribute: %s\n", err.Error())
			}
			mm := &pb.MumbleMurmer{}
			proto.Unmarshal(body, mm)
			m.ReceiveMurmer(mm)
		})
	*/
}

func (g *grapevineServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/gossip", g.gossip)
	mux.HandleFunc("/gossip/searchresult", g.gossip)
	mux.HandleFunc("/data/invite", g.gossip)
	mux.HandleFunc("/data/change/owner", g.gossip)
	mux.HandleFunc("/data/change/data", g.gossip)
	mux.HandleFunc("/data/leave", g.gossip)

	quicConf := &quic.Config{}

	pool, err := x509.SystemCertPool()
	if err != nil {
		log.Fatal(err)
	}

	ip := "localhost"
	port := 8911

	for {
		addr := fmt.Sprintf("%s:%d", ip, port)

		server := http3.Server{
			Handler:    mux,
			Addr:       addr,
			QuicConfig: quicConf,
			TLSConfig: &tls.Config{
				RootCAs:            pool,
				InsecureSkipVerify: true,
			},
		}

		fmt.Printf("Trying to listening on %v\n", server.Addr)
		err = server.ListenAndServe()
		if err != nil {
			if strings.Contains(err.Error(), "bind: address already in use") {
				fmt.Println("port already used, incrementing")
				port++
			} else {
				return err
			}
		}
	}
}

func StartClient() {
	err := NewServer().Start()
	if err != nil {
		fmt.Println(err)
	}

}

func main() {
	StartClient()
}
