package client

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"path"
	"runtime"

	"github.com/golang/protobuf/proto"
	pb "github.com/hoyle1974/grapevine/proto"
	"github.com/hoyle1974/grapevine/services"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

type GrapevineServer interface {
	Start(net.IP) (int, error)
	GetIp() net.IP
	GetPort() int
	SetGossip(gossip Gossip)
}

type grapevineServer struct {
	ctx  CallCtx
	ip   net.IP
	port int
	g    Gossip
}

func (g *grapevineServer) GetIp() net.IP {
	return g.ip
}

func (g *grapevineServer) GetPort() int {
	return g.port
}

func (g *grapevineServer) SetGossip(gossip Gossip) {
	g.g = gossip
}

func NewServer(ctx CallCtx) GrapevineServer {
	return &grapevineServer{ctx: ctx.NewCtx("server")}
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
	log := g.ctx.NewCtx("gossip")

	if req.Method == "GET" {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Printf("error reading body while handling /distribute: %s", err.Error())
		}
		gr := &pb.GossipRequest{}
		proto.Unmarshal(body, gr)

		log.Printf("Gossip (via GET): %v", gr.Gossip)

		for _, v := range gr.Gossip {
			s := v.GetSearch()
			if s != nil {
				ip := net.ParseIP(s.Requestor.Address.IpAddress)
				port := s.Requestor.Address.Port
				g.g.AddServer(services.NewServerAddress(ip, port))
			}
		}
	}
	if req.Method == "POST" {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Printf("error reading body while handling /distribute: %s", err.Error())
		}
		gr := &pb.GossipResponse{}
		proto.Unmarshal(body, gr)
		log.Printf("Gossip (via POST): %v", gr.Gossip)

		for _, v := range gr.Gossip {
			s := v.GetSearch()
			if s != nil {
				ip := net.ParseIP(s.Requestor.Address.IpAddress)
				port := s.Requestor.Address.Port
				g.g.AddServer(services.NewServerAddress(ip, port))
			}
		}
	}
}

func (g *grapevineServer) isPortAvailable(ip net.IP, port int) bool {
	log := g.ctx.NewCtx("isPortAvailable")

	addr := net.UDPAddr{
		IP:   ip,
		Port: port,
	}
	conn, err := net.ListenUDP("udp", &addr)

	if err != nil {
		log.Warn().Msgf("Can't listen on port %d: %s", port, err)
		return false
	}

	conn.Close()
	log.Info().Msgf("TCP Port %v is available", port)
	return true
}

var certPath string

func init() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("Failed to get current frame")
	}

	certPath = path.Dir(filename)
}

func GetCertificatePaths() (string, string) {
	return path.Join(certPath, "cert.pem"), path.Join(certPath, "priv.key")
}

func (g *grapevineServer) Start(ip net.IP) (int, error) {
	log := g.ctx.NewCtx("Start")

	g.ip = ip

	mux := http.NewServeMux()
	mux.HandleFunc("/gossip", g.gossip)
	// mux.HandleFunc("/data/invite", g.gossip)
	// mux.HandleFunc("/data/change/owner", g.gossip)
	// mux.HandleFunc("/data/change/data", g.gossip)
	// mux.HandleFunc("/data/leave", g.gossip)

	quicConf := &quic.Config{}

	g.port = 8911

	for g.isPortAvailable(ip, g.port) == false {
		g.port++
	}

	addr := fmt.Sprintf("%s:%d", "", g.port)

	server := http3.Server{
		Handler:    mux,
		Addr:       addr,
		QuicConfig: quicConf,
	}

	log.Info().Msgf("Listening on %v", server.Addr)
	go func() {
		err := server.ListenAndServeTLS(GetCertificatePaths())
		if err != nil {
			panic(err)
		}
	}()

	return g.port, nil
}
