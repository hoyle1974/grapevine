package client

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	protoc "github.com/golang/protobuf/proto"
	pb "github.com/hoyle1974/grapevine/proto"
	"github.com/hoyle1974/grapevine/services"
	"github.com/rs/zerolog/log"
)

type SharedDataManager interface {
	Serve(s SharedData)
	JoinShare(s SharedData)
	LeaveShare(s SharedData)
	Invite(s SharedData, recipient services.UserContact, as string) bool
}

type sharedDataManager struct {
	lock        sync.Mutex
	data        map[SharedDataId]SharedData
	clientCache GrapevineClientCache
}

func NewSharedDataManager(clientCache GrapevineClientCache) SharedDataManager {
	return &sharedDataManager{clientCache: clientCache}
}

func (sdm *sharedDataManager) Serve(s SharedData) {
	sdm.lock.Lock()
	defer sdm.lock.Unlock()

	// Add this shared data to our system
	sdm.data[s.GetId()] = s
}

func (sdm *sharedDataManager) JoinShare(s SharedData) {
	sdm.lock.Lock()
	defer sdm.lock.Unlock()

	todo

}

func (sdm *sharedDataManager) LeaveShare(s SharedData) {
	sdm.lock.Lock()
	defer sdm.lock.Unlock()

	todo

}

func (sdm *sharedDataManager) Invite(s SharedData, recipient services.UserContact, as string) bool {
	sdm.lock.Lock()
	defer sdm.lock.Unlock()

	// Tell the contact about this shared data, inviting them to it

	client := sdm.clientCache.GetClient(recipient).GetClient()

	invite := pb.SharedDataInvite{
		SharedDataId: string(s.GetId()),
		Creator:      s.GetCreator().GetPB(),
	}

	b, err := protoc.Marshal(invite)
	if err != nil {
		log.Error().Err(err).Msg("Can't unmarshal")
		return false
	}

	resp, err := client.Post(fmt.Sprintf("https://%s/shareddata", recipient.GetURL()), "grpc-message-type", bytes.NewReader(b))
	if err != nil {
		log.Error().Err(err).Msg("\tTried to post but got error")
		return false
	} else {
		log.Info().Msgf("\tResponse: %v", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("\tError reading body of response")
		return false
	}

	gresp := pb.SharedDataInviteResponse{}
	err = protoc.Unmarshal(b, &gresp)
	if err != nil {
		log.Error().Err(err).Msg("\tError unmarshaling response to our gossip")
		return false
	}

	return true
}
