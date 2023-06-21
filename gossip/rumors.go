package gossip

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hoyle1974/grapevine/common"
	"github.com/hoyle1974/grapevine/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Rumor interface {
	GetRumorId() uuid.UUID
	GetExpiry() time.Time
	GetCreator() common.Contact
	IsExpired() bool
	ToProtobuf() *proto.Gossip
	String() string
}

type rumor struct {
	rumorId uuid.UUID
	expiry  time.Time
	creator common.Contact
}

func NewRumor(rumorId uuid.UUID, expiry time.Time, creatorAccountId common.AccountId, creatorAddr common.Address) rumor {
	return rumor{
		rumorId: rumorId,
		expiry:  expiry,
		creator: common.NewContact(creatorAccountId, creatorAddr.Ip, creatorAddr.Port),
	}
}

func (r rumor) String() string {
	return fmt.Sprintf("Id(%v) Expired(%v) Creator(%v)",
		r.rumorId.String(),
		r.expiry.String(),
		r.creator.String(),
	)
}

func (r rumor) GetRumorId() uuid.UUID {
	return r.rumorId
}

func (r rumor) GetExpiry() time.Time {
	return r.expiry
}

func (r rumor) GetCreator() common.Contact {
	return r.creator
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
		SearchId:  r.rumorId.String(),
		Query:     r.query,
		Requestor: r.creator.ToPB(),
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
	ctx    common.CallCtx
	rumors []Rumor
}

func NewRumors(ctx common.CallCtx) Rumors {
	return &rumors{ctx: ctx}
}

func (r *rumors) AddRumor(rumor Rumor) {
	//log := r.ctx.NewCtx("AddRumor")

	r.lock.Lock()
	defer r.lock.Unlock()

	for _, rr := range r.rumors {
		if rumor.GetRumorId() == rr.GetRumorId() {
			return
		}
	}

	//log.Info().Msgf("Adding rumor: %v", rumor)
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

	return &proto.GossipRequest{Gossip: toGossip}
}
