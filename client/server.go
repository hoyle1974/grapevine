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
	"github.com/rs/zerolog/log"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

type GrapevineServer interface {
	Start(net.IP) (int, error)
	GetIp() net.IP
	GetPort() int
}

type grapevineServer struct {
	ip   net.IP
	port int
}

func (g *grapevineServer) GetIp() net.IP {
	return g.ip
}

func (g *grapevineServer) GetPort() int {
	return g.port
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
	fmt.Println("***** receive a gossip")
	if req.Method == "GET" {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			fmt.Printf("error reading body while handling /distribute: %s\n", err.Error())
		}
		gr := &pb.GossipRequest{}
		proto.Unmarshal(body, gr)

		fmt.Printf("Gossip: %v\n", gr.Gossip)
	}
	if req.Method == "POST" {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			fmt.Printf("error reading body while handling /distribute: %s\n", err.Error())
		}
		gr := &pb.GossipResponse{}
		proto.Unmarshal(body, gr)
		fmt.Printf("Gossip: %v\n", gr.Gossip)
	}
}

func isPortAvailable(ip net.IP, port int) bool {

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
	g.ip = ip

	mux := http.NewServeMux()
	mux.HandleFunc("/gossip", g.gossip)
	// mux.HandleFunc("/data/invite", g.gossip)
	// mux.HandleFunc("/data/change/owner", g.gossip)
	// mux.HandleFunc("/data/change/data", g.gossip)
	// mux.HandleFunc("/data/leave", g.gossip)

	quicConf := &quic.Config{}

	// pool, err := x509.SystemCertPool()
	// if err != nil {
	// 	return 0, err
	// }

	g.port = 8911

	for isPortAvailable(ip, g.port) == false {
		g.port++
	}

	addr := fmt.Sprintf("%s:%d", ip.String(), g.port)

	server := http3.Server{
		Handler:    mux,
		Addr:       addr,
		QuicConfig: quicConf,
		// TLSConfig: &tls.Config{
		// 	ServerName:            "localhost",
		// 	RootCAs:               pool,
		// 	InsecureSkipVerify:    true,
		// 	VerifyConnection:      nil,
		// 	VerifyPeerCertificate: nil,
		// },
	}

	log.Info().Msgf("Listening on %v\n", server.Addr)
	go func() {
		err := server.ListenAndServeTLS(GetCertificatePaths())
		if err != nil {
			panic(err)
		}
	}()

	return g.port, nil
}
