package client

import (
	"io"
	"net/http"
	"sync"

	"github.com/golang/protobuf/proto"
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
	OnSharedDataRequest(writer http.ResponseWriter, req *http.Request)
}

type sharedDataManager struct {
	lock        sync.Mutex
	ctx         CallCtx
	myself      common.Myself
	data        map[SharedDataId]SharedDataProxy
	clientCache GrapevineClientCache
}

func NewSharedDataManager(ctx CallCtx, myself common.Myself, clientCache GrapevineClientCache) SharedDataManager {
	return &sharedDataManager{
		clientCache: clientCache,
		myself:      myself,
		ctx:         ctx.NewCtx("SharedDataManager"),
		data:        make(map[SharedDataId]SharedDataProxy),
	}
}

func (sdm *sharedDataManager) OnSharedDataRequest(writer http.ResponseWriter, req *http.Request) {
	sdm.ctx.Warn().Msgf("OnSharedDataRequest - %v", req.RequestURI)

	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Error().Err(err).Msg("error reading body while handling /distribute")
		writer.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	var resp proto.Message = nil

	sdm.lock.Lock()
	defer sdm.lock.Unlock()

	switch req.RequestURI {
	case "/shareddata/invite":
		req := &pb.SharedDataInvite{}
		proto.Unmarshal(body, req)
		// We were invite to this shared data, make sure the CB knows

		// If we accept then we will create the object

		sd := NewSharedData(common.NewContactFromPB(req.Creator), SharedDataId(req.SharedDataId))
		sd.SetMe(req.As)

		sdm.data[sd.GetId()] = NewSharedDataProxy(sd, sdm)

		resp = &pb.SharedDataInviteResponse{}
	case "/shareddata/set":
		req := &pb.SharedDataSet{}
		proto.Unmarshal(body, req)

		id := SharedDataId(req.SharedDataId)
		sdm.data[id].GetOrigin().Set(req.Key, req.Value)

		resp = &pb.SharedDataSetResponse{}
	case "/shareddata/append":
		req := &pb.SharedDataAppend{}
		proto.Unmarshal(body, req)

		id := SharedDataId(req.SharedDataId)
		sdm.data[id].GetOrigin().Append(req.Key, req.Value)

		resp = &pb.SharedDataAppendResponse{}
	case "/shareddata/changeowner":
		req := &pb.SharedDataChangeOwner{}
		proto.Unmarshal(body, req)

		id := SharedDataId(req.SharedDataId)
		sdm.data[id].GetOrigin().ChangeDataOwner(req.Key, req.Owner)

		resp = &pb.SharedDataChangeOwnerResponse{}
	default:
		sdm.ctx.Error().Msgf("Unsupported shared data command: %s", req.RequestURI)
	}

	if resp != nil {
		body, err = proto.Marshal(resp)
		if err != nil {
			log.Error().Err(err).Msg("error writing response")
			writer.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		writer.WriteHeader(http.StatusOK)
		writer.Write(body)
	} else {
		writer.WriteHeader(http.StatusServiceUnavailable)
	}

}

func (sdm *sharedDataManager) GetMe() common.Contact {
	return sdm.myself.GetMe()
}

func (sdm *sharedDataManager) Serve(s SharedData) SharedData {
	sdm.ctx.Info().Msg("Server")
	sdm.lock.Lock()
	defer sdm.lock.Unlock()

	proxy := NewSharedDataProxy(s, sdm)

	// Add this shared data to our system
	sdm.data[s.GetId()] = proxy

	return proxy
}

func (sdm *sharedDataManager) JoinShare(s SharedData) {
	sdm.ctx.Info().Msg("JoinShare")
	sdm.lock.Lock()
	defer sdm.lock.Unlock()

	// @TODO

}

func (sdm *sharedDataManager) LeaveShare(s SharedData) {
	sdm.ctx.Info().Msg("LeaveShare")

	sdm.lock.Lock()
	defer sdm.lock.Unlock()

	// @TODO

}

func (sdm *sharedDataManager) Invite(s SharedData, recipient common.Contact, as string) bool {
	sdm.ctx.Info().Msg("Invite")

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
