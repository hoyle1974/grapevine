package client

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"path"
	"runtime"

	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	pb "github.com/hoyle1974/grapevine/proto"
	"github.com/hoyle1974/grapevine/services"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

type GrapevineListener interface {
	Listen(net.IP) (int, error)
	GetIp() net.IP
	GetPort() int
	SetGossip(gossip Gossip)
}

type grapevineListener struct {
	ctx  CallCtx
	ip   net.IP
	port int
	g    Gossip
}

func (g *grapevineListener) GetIp() net.IP {
	return g.ip
}

func (g *grapevineListener) GetPort() int {
	return g.port
}

func (g *grapevineListener) SetGossip(gossip Gossip) {
	g.g = gossip
}

func NewGrapevineListener(ctx CallCtx) GrapevineListener {
	return &grapevineListener{ctx: ctx.NewCtx("server")}
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

func (g *grapevineListener) onGossip(writer http.ResponseWriter, req *http.Request) {
	log := g.ctx.NewCtx("onGossip")

	log.Info().Msg("\tReceive")

	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Error().Err(err).Msg("error reading body while handling /distribute")
		writer.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	gr := &pb.GossipRequest{}
	proto.Unmarshal(body, gr)

	log.Info().Msgf("\tContains %v messages", len(gr.Gossip))

	for _, gg := range gr.Gossip {
		search := gg.GetSearch()
		if search != nil {

			rumorId, err := uuid.Parse(search.SearchId)
			if err != nil {
				log.Warn().Err(err).Msgf("Couldn't parse: %v", search.SearchId)
				continue
			}

			rumor := NewSearchRumor(NewRumor(
				rumorId,
				gg.EndOfLife.AsTime(),
				services.NewAccountId(search.Requestor.AccountId),
				services.NewServerAddress(net.ParseIP(search.Requestor.Address.IpAddress), search.Requestor.Address.Port),
			), search.Query)

			go g.g.AddToGossip(rumor)
		} else {
			log.Warn().Msgf("Unknown gossip: %v", gg)
		}
	}

	body, err = proto.Marshal(&pb.GossipResponse{})
	if err != nil {
		log.Error().Err(err).Msg("error writing response")
		writer.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(body)
}

func (g *grapevineListener) isPortAvailable(ip net.IP, port int) bool {
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

func (g *grapevineListener) Listen(ip net.IP) (int, error) {
	log := g.ctx.NewCtx("Start")

	g.ip = ip

	mux := http.NewServeMux()
	mux.HandleFunc("/gossip", g.onGossip)
	// mux.HandleFunc("/data/invite", g.gossip)
	// mux.HandleFunc("/data/change/owner", g.gossip)
	// mux.HandleFunc("/data/change/data", g.gossip)
	// mux.HandleFunc("/data/leave", g.gossip)

	quicConf := &quic.Config{}

	g.port = 8911

	for !g.isPortAvailable(ip, g.port) {
		g.port++
	}

	addr := fmt.Sprintf("%s:%d", ip, g.port)

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
