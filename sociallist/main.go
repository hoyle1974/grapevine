package main

import (
	"context"

	"github.com/hoyle1974/grapevine/common"
	"github.com/hoyle1974/grapevine/microservice"
	pb "github.com/hoyle1974/grapevine/proto"
	"github.com/hoyle1974/grapevine/services"
)

type server struct {
	pb.UnimplementedSocialListServiceServer
	appCtx services.AppCtx
}

func PBSocialListTypeToSocialListType(s pb.SocialListType) services.SocialListType {
	var slt services.SocialListType

	switch s {
	case pb.SocialListType_BLOCKED:
		slt = services.SocialListType_BLOCKED
	case pb.SocialListType_FOLLOW:
		slt = services.SocialListType_FOLLOWS
	}
	return slt
}

func (s *server) GetSocialList(ctx context.Context, in *pb.GetSocialListRequest) (*pb.GetSocialListResponse, error) {
	items, err := services.GetSocialList(
		s.appCtx,
		common.NewAccountId(in.GetUserId()),
		PBSocialListTypeToSocialListType(in.GetListType()),
	)
	if err != nil {
		return &pb.GetSocialListResponse{Error: microservice.ErrToProto(err)}, err
	}

	entities := make([]string, 0)
	for _, entity_id := range items {
		entities = append(entities, entity_id.String())
	}

	return &pb.GetSocialListResponse{Ids: entities}, nil
}

func (s *server) AddToSocialList(ctx context.Context, in *pb.AddToSocialListRequest) (*pb.AddToSocialListResponse, error) {

	err := services.AddToSocialList(
		s.appCtx,
		common.NewAccountId(in.UserId),
		PBSocialListTypeToSocialListType(in.SocialListType),
		common.NewAccountId(in.IdToAdd),
	)
	if err != nil {
		return &pb.AddToSocialListResponse{Error: microservice.ErrToProto(err)}, err
	}

	return &pb.AddToSocialListResponse{}, nil
}

func (s *server) RemoveFromSocialList(ctx context.Context, in *pb.RemoveFromSocialListRequest) (*pb.RemoveFromSocialListResponse, error) {
	err := services.RemoveFromSocialList(
		s.appCtx,
		common.NewAccountId(in.UserId),
		in.GetSocialListType().String(),
		common.NewAccountId(in.IdToRemove),
	)
	if err != nil {
		return &pb.RemoveFromSocialListResponse{Error: microservice.ErrToProto(err)}, err
	}

	return &pb.RemoveFromSocialListResponse{}, nil
}

func register(appCtx services.AppCtx) {
	pb.RegisterSocialListServiceServer(appCtx.Server, &server{appCtx: appCtx})
}

func main() {
	microservice.Start("sociallist", register)
}
