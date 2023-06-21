package shareddata

import (
	"io"
	"net/http"
	"sync"

	"github.com/hoyle1974/grapevine/client"
	"github.com/hoyle1974/grapevine/common"
	pb "github.com/hoyle1974/grapevine/proto"
	proto "google.golang.org/protobuf/proto"
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
	ctx         common.CallCtx
	myself      common.Myself
	cb          ClientCallback
	data        map[SharedDataId]SharedDataProxy
	clientCache client.GrapevineClientCache
}

func NewSharedDataManager(ctx common.CallCtx, myself common.Myself, cb ClientCallback, clientCache client.GrapevineClientCache) SharedDataManager {
	return &sharedDataManager{
		clientCache: clientCache,
		myself:      myself,
		cb:          cb,
		ctx:         ctx.NewCtx("SharedDataManager"),
		data:        make(map[SharedDataId]SharedDataProxy),
	}
}

func (sdm *sharedDataManager) OnSharedDataRequest(writer http.ResponseWriter, req *http.Request) {
	log := sdm.ctx.NewCtx("OnSharedDataRequest")
	//sdm.ctx.Info().Str("uri", req.RequestURI).Msg("request")

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
		creator := common.NewContactFromPB(req.Creator)

		// If we accept then we will create the object
		if sdm.cb.OnInvited(SharedDataId(req.SharedDataId), req.As, creator) {
			sd := NewSharedData(creator, SharedDataId(req.SharedDataId))
			sd.SetMe(req.As)
			proxy := NewSharedDataProxy(sd, sdm)
			proxy.AddInvitee(sdm.GetMe(), req.As)
			sdm.data[sd.GetId()] = proxy
			resp = &pb.SharedDataInviteResponse{Accepted: true}

			sdm.cb.OnSharedDataAvailable(proxy)
		} else {
			resp = &pb.SharedDataInviteResponse{Accepted: false}
		}
	case "/shareddata/sendstate":
		req := &pb.SharedDataSendState{}
		proto.Unmarshal(body, req)

		id := SharedDataId(req.SharedDataId)
		sd := sdm.data[id]

		for key, value := range req.Data {
			sd.Create(key, value.Value, value.Owner, value.Visbility)
		}

		for key, value := range req.Listeners {
			sd.AddInvitee(common.NewContactFromPB(value), key)
		}
		resp = &pb.SharedDataSendStateResponse{}
	case "/shareddata/create":
		req := &pb.SharedDataCreate{}
		proto.Unmarshal(body, req)

		id := SharedDataId(req.SharedDataId)

		sdm.data[id].GetOrigin().Create(req.Key, req.Value, req.Owner, req.Visibility)

		resp = &pb.SharedDataCreateResponse{}
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
		log.Error().Msgf("Unsupported shared data command: %s", req.RequestURI)
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
	// log := sdm.ctx.NewCtx("Serve")
	sdm.lock.Lock()
	defer sdm.lock.Unlock()

	proxy := NewSharedDataProxy(s, sdm)
	proxy.AddInvitee(sdm.GetMe(), s.GetMe())

	// Add this shared data to our system
	sdm.data[s.GetId()] = proxy

	return proxy
}

func (sdm *sharedDataManager) JoinShare(s SharedData) {
	// log := sdm.ctx.NewCtx("JoinShare")
	sdm.lock.Lock()
	defer sdm.lock.Unlock()

	// @TODO

}

func (sdm *sharedDataManager) LeaveShare(s SharedData) {
	// log := sdm.ctx.NewCtx("LeaveShare")

	sdm.lock.Lock()
	defer sdm.lock.Unlock()

	// @TODO

}

func (sdm *sharedDataManager) Invite(s SharedData, recipient common.Contact, as string) bool {
	log := sdm.ctx.NewCtx("Invite")

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

	if gresp.Accepted {
		proxy := sdm.data[s.GetId()]
		sdm.ctx.Info().Msgf("Add Invitee %v", recipient)
		proxy.AddInvitee(recipient, as)

		sdm.ctx.Info().Msgf("Send State %v", recipient)
		proxy.SendStateTo(recipient)

		sdm.ctx.Info().Msgf("Invite Accepted by %v", recipient)
		go sdm.cb.OnInviteAccepted(proxy, recipient)
		return true
	}

	return false
}
