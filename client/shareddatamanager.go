package client

import (
	"sync"

	"github.com/hoyle1974/grapevine/common"
	pb "github.com/hoyle1974/grapevine/proto"
	"github.com/rs/zerolog/log"
)

type SharedDataManager interface {
	GetMe() common.Contact
	Serve(s SharedData) SharedData
	JoinShare(s SharedData)
	LeaveShare(s SharedData)
	Invite(s SharedData, recipient common.Contact, as string) bool
}

type sharedDataManager struct {
	lock        sync.Mutex
	myself      common.Myself
	data        map[SharedDataId]SharedDataProxy
	clientCache GrapevineClientCache
}

func NewSharedDataManager(myself common.Myself, clientCache GrapevineClientCache) SharedDataManager {
	return &sharedDataManager{clientCache: clientCache, myself: myself}
}

func (sdm *sharedDataManager) GetMe() common.Contact {
	return sdm.myself.GetMe()
}

func (sdm *sharedDataManager) Serve(s SharedData) SharedData {
	sdm.lock.Lock()
	defer sdm.lock.Unlock()

	proxy := NewSharedDataProxy(s, sdm)

	// Add this shared data to our system
	sdm.data[s.GetId()] = proxy

	return proxy
}

func (sdm *sharedDataManager) JoinShare(s SharedData) {
	sdm.lock.Lock()
	defer sdm.lock.Unlock()

	// @TODO

}

func (sdm *sharedDataManager) LeaveShare(s SharedData) {
	sdm.lock.Lock()
	defer sdm.lock.Unlock()

	// @TODO

}

func (sdm *sharedDataManager) Invite(s SharedData, recipient common.Contact, as string) bool {
	sdm.lock.Lock()
	defer sdm.lock.Unlock()

	// Tell the contact about this shared data, inviting them to it
	invite := pb.SharedDataInvite{
		SharedDataId: string(s.GetId()),
		Creator:      s.GetCreator().ToPB(),
		As:           as,
	}

	gresp := pb.SharedDataInviteResponse{}

	err := sdm.clientCache.POST(recipient.Address, "/shareddata/invite", &invite, &gresp)
	if err != nil {
		log.Error().Err(err).Msg("Can't unmarshal")
		return false
	}

	proxy := sdm.data[s.GetId()]
	proxy.AddInvitee(recipient, as)

	return true
}
