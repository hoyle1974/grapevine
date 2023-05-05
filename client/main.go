package main

import (
	"context"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/hoyle1974/grapevine/microservice"
	pb "github.com/hoyle1974/grapevine/proto"
	"github.com/hoyle1974/grapevine/services"
)

type server struct {
	pb.UnimplementedGrapevineServiceServer
	appCtx    services.AppCtx
	grapevine Grapevine
}

func (s *server) Gossip(ctx context.Context, in *pb.GossipRequest) (*pb.GossipResponse, error) {

	for _, gossip := range in.GetGossip() {
		if gossip.EndOfLife.AsTime().Before(time.Now()) {
			// Gossip still valid
			if search := gossip.GetSearch(); s != nil {
				requestor := services.UserContact{
					AccountID: services.NewAccountId(search.Requestor.GetAccountId()),
					Ip:        net.ParseIP(search.GetRequestor().GetAddress().GetIpAddress()),
					Port:      search.GetRequestor().GetAddress().GetPort(),
					Time:      time.Now(),
				}
				go s.grapevine.OnGossipSearch(s.appCtx, requestor, search.Query, search)
			}
		}
	}

	return &pb.GossipResponse{}, nil
}

func (s *server) SharedInvitation(ctx context.Context, in *pb.SharedInvitationRequest) (*pb.SharedInvitationResponse, error) {
	return &pb.SharedInvitationResponse{}, nil
}

func (s *server) ChangeDataOwner(ctx context.Context, in *pb.ChangeDataOwnerRequest) (*pb.ChangeDataOwnerResponse, error) {
	return &pb.ChangeDataOwnerResponse{}, nil
}

func (s *server) ChangeData(ctx context.Context, in *pb.ChangeDataRequest) (*pb.ChangeDataResponse, error) {
	return &pb.ChangeDataResponse{}, nil
}

func (s *server) LeaveSharedData(ctx context.Context, in *pb.LeaveSharedDataRequest) (*pb.LeaveSharedDataResponse, error) {
	return &pb.LeaveSharedDataResponse{}, nil
}

// Client callbacks
func (s *server) OnSearchQuery(query string) bool {
	return query == "mvp/tictactoe/v1"
}

func (s *server) OnSharedInvitation() {}
func (s *server) OnChangeOwner()      {}
func (s *server) OnChangeData()       {}
func (s *server) OnLeaveSharedData()  {}

func register(appCtx services.AppCtx) {
	s := &server{appCtx: appCtx}
	s.grapevine = NewGrapevine(s, services.NewAccountId(uuid.NewString()), appCtx.GetAddr())
	pb.RegisterGrapevineServiceServer(appCtx.Server, s)
}

func tempGrapevineStart() {
	microservice.Start("grapevine", register)
}

// ------------------------------------

func StartServer() {

}
