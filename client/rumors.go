package client

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hoyle1974/grapevine/proto"
	"github.com/hoyle1974/grapevine/services"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Rumor interface {
	GetRumorId() uuid.UUID
	GetExpiry() time.Time
	GetCreatorAccountId() services.AccountId
	GetCreatorAddress() services.ServerAddress
	IsExpired() bool
	ToProtobuf() *proto.Gossip
	String() string
}

type rumor struct {
	rumorId          uuid.UUID
	expiry           time.Time
	creatorAccountId services.AccountId
	creatorAddr      services.ServerAddress
}

func NewRumor(rumorId uuid.UUID, expiry time.Time, creatorAccountId services.AccountId, creatorAddr services.ServerAddress) rumor {
	return rumor{
		rumorId:          rumorId,
		expiry:           expiry,
		creatorAccountId: creatorAccountId,
		creatorAddr:      creatorAddr,
	}
}

func (r rumor) String() string {
	return fmt.Sprintf("Id(%v) Expired(%v) Creator(%v,%v)",
		r.rumorId.String(),
		r.expiry.String(),
		r.creatorAccountId.String(),
		r.creatorAddr.String(),
	)
}

func (r rumor) GetRumorId() uuid.UUID {
	return r.rumorId
}

func (r rumor) GetExpiry() time.Time {
	return r.expiry
}

func (r rumor) GetCreatorAccountId() services.AccountId {
	return r.creatorAccountId
}

func (r rumor) GetCreatorAddress() services.ServerAddress {
	return r.creatorAddr
}

func (r rumor) IsExpired() bool {
	return time.Now().After(r.expiry)
}

type SearchRumor struct {
	rumor
	query string
}

func NewSearchRumor(rumor rumor, query string) SearchRumor {
	return SearchRumor{rumor: rumor, query: query}
}

func (r SearchRumor) String() string {
	return fmt.Sprintf("%v Query(%v)",
		r.rumor.String(),
		r.query,
	)
}

func (r SearchRumor) GetQuery() string {
	return r.query
}

func (r SearchRumor) ToProtobuf() *proto.Gossip {
	search := &proto.Search{
		SearchId: r.rumorId.String(),
		Query:    r.query,
		Requestor: &proto.Contact{
			AccountId: string(r.creatorAccountId.String()),
			Address: &proto.ClientAddress{
				IpAddress: r.creatorAddr.GetIp().String(),
				Port:      int32(r.creatorAddr.GetPort()),
			},
		},
	}

	return &proto.Gossip{
		EndOfLife:   timestamppb.New(r.expiry),
		GossipUnion: &proto.Gossip_Search{Search: search},
	}
}

type Rumors interface {
	AddRumor(rumor Rumor)
	GetProtobuf() *proto.GossipRequest
}

type rumors struct {
	lock   sync.Mutex
	ctx    CallCtx
	rumors []Rumor
}

func NewRumors(ctx CallCtx) Rumors {
	return &rumors{ctx: ctx}
}

func (r *rumors) AddRumor(rumor Rumor) {
	log := r.ctx.NewCtx("AddRumor")

	r.lock.Lock()
	defer r.lock.Unlock()

	for _, rr := range r.rumors {
		if rumor.GetRumorId() == rr.GetRumorId() {
			return
		}
	}

	log.Info().Msgf("Adding rumor: %v", rumor)
	r.rumors = append(r.rumors, rumor)
}

func (r *rumors) GetProtobuf() *proto.GossipRequest {
	log := r.ctx.NewCtx("GetProtobuf")

	r.lock.Lock()
	defer r.lock.Unlock()

	toGossip := []*proto.Gossip{}

	newRumors := []Rumor{}
	for _, rr := range r.rumors {
		if !rr.IsExpired() {
			newRumors = append(newRumors, rr)
			toGossip = append(toGossip, rr.ToProtobuf())
		} else {
			log.Info().Msgf("Removing expired rumor: %v", rr.GetRumorId())
		}
	}

	r.rumors = newRumors

	log.Debug().Msgf("Returning gossip request with %d reuors", len(toGossip))
	return &proto.GossipRequest{Gossip: toGossip}
}
