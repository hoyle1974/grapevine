package main

import (
	"context"
	"net"

	"github.com/hoyle1974/grapevine/common"
	"github.com/hoyle1974/grapevine/microservice"
	pb "github.com/hoyle1974/grapevine/proto"
	"github.com/hoyle1974/grapevine/services"
)

type server struct {
	pb.UnimplementedAuthServiceServer
	appCtx services.AppCtx
}

// func UserContactsToPB(contacts []services.UserContact) []*pb.UserContact {
// 	out := make([]*pb.UserContact, len(contacts))
// 	for _, contact := range contacts {
// 		c := pb.UserContact{
// 			UserId: contact.AccountID.String(),
// 			ClientAddress: &pb.ClientAddress{
// 				IpAddress: contact.Ip.String(),
// 				Port:      contact.Port,
// 			},
// 		}
// 		out = append(out, &c)
// 	}

// 	return out
// }

func (s *server) Auth(ctx context.Context, in *pb.AuthRequest) (*pb.AuthResponse, error) {

	accountId, err := services.Auth(
		s.appCtx,
		in.GetUsername(),
		in.GetPassword(),
		net.ParseIP(in.GetClientAddress().GetIpAddress()),
		in.GetClientAddress().GetPort(),
	)
	if err != nil {
		return &pb.AuthResponse{Error: microservice.ErrToProto(err)}, err
	}

	followsIds, err := services.GetSocialList(s.appCtx, accountId, services.SocialListType_FOLLOWS)
	if err != nil {
		return &pb.AuthResponse{Error: microservice.ErrToProto(err)}, err
	}
	followingIds, err := services.GetSocialList(s.appCtx, accountId, services.SocialListType_FOLLOWING)
	if err != nil {
		return &pb.AuthResponse{Error: microservice.ErrToProto(err)}, err
	}
	blockedIds, err := services.GetSocialList(s.appCtx, accountId, services.SocialListType_BLOCKED)
	if err != nil {
		return &pb.AuthResponse{Error: microservice.ErrToProto(err)}, err
	}

	follows, err := services.GetUserContacts(s.appCtx, followsIds)
	if err != nil {
		return &pb.AuthResponse{Error: microservice.ErrToProto(err)}, err
	}
	following, err := services.GetUserContacts(s.appCtx, followingIds)
	if err != nil {
		return &pb.AuthResponse{Error: microservice.ErrToProto(err)}, err
	}

	return &pb.AuthResponse{
		Message:   "Hello " + in.GetUsername(),
		UserId:    accountId.String(),
		Blocked:   common.AccountIdsToStrings(blockedIds),
		Follows:   common.ContactsToPB(follows),
		Following: common.ContactsToPB(following),
	}, nil

}

func register(appCtx services.AppCtx) {
	pb.RegisterAuthServiceServer(appCtx.Server, &server{appCtx: appCtx})
}

func main() {
	microservice.Start("auth", register)
}
