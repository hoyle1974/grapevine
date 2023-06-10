package client

import (
	"fmt"
	"sync"

	"github.com/hoyle1974/grapevine/common"
	pb "github.com/hoyle1974/grapevine/proto"
)

type SharedDataProxy interface {
	SharedData
	AddInvitee(recipient common.Contact, as string)
}

func NewSharedDataProxy(origin SharedData, sdm *sharedDataManager) SharedDataProxy {

	if origin.IsProxy() {
		return nil
	}

	return &sharedDataProxy{
		origin:   origin,
		sdm:      sdm,
		invities: make(map[string]common.Contact),
	}
}

func (p *sharedDataProxy) AddInvitee(recipient common.Contact, as string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.invities[as] = recipient
}

type sharedDataProxy struct {
	lock     sync.Mutex
	origin   SharedData
	sdm      *sharedDataManager
	invities map[string]common.Contact
}

func (p *sharedDataProxy) IsProxy() bool {
	return true
}

func (p *sharedDataProxy) GetCreator() common.Contact {
	return p.origin.GetCreator()
}

func (p *sharedDataProxy) GetId() SharedDataId {
	return p.origin.GetId()
}

func (p *sharedDataProxy) Create(key string, value interface{}, owner string, visibility string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.origin.Create(key, value, owner, visibility)

	req := pb.SharedDataCreate{
		SharedDataId: string(p.origin.GetId()),
		Originator:   p.sdm.GetMe().ToPB(),
		Key:          key,
		Value:        fmt.Sprintf("%v", value),
		Owner:        owner,
		Visibility:   visibility,
	}
	resp := pb.SharedDataCreateResponse{}
	for _, value := range p.invities {
		p.sdm.clientCache.POST(value.Address, "/shareddata/create", &req, &resp)
	}
}

func (p *sharedDataProxy) Get(key string) interface{} {
	return p.origin.Get(key)
}

func (p *sharedDataProxy) Set(key string, value interface{}) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.origin.Set(key, value)

	req := pb.SharedDataSet{
		SharedDataId: string(p.origin.GetId()),
		Originator:   p.sdm.GetMe().ToPB(),
		Key:          key,
		Value:        fmt.Sprintf("%v", value),
	}
	resp := pb.SharedDataSetResponse{}
	for _, value := range p.invities {
		p.sdm.clientCache.POST(value.Address, "/shareddata/set", &req, &resp)
	}
}

func (p *sharedDataProxy) Append(key string, value interface{}) {
	p.lock.Lock()
	defer p.lock.Unlock()

	req := pb.SharedDataAppend{
		SharedDataId: string(p.origin.GetId()),
		Originator:   p.sdm.GetMe().ToPB(),
		Key:          key,
		Value:        fmt.Sprintf("%v", value),
	}
	resp := pb.SharedDataAppendResponse{}
	for _, value := range p.invities {
		p.sdm.clientCache.POST(value.Address, "/shareddata/append", &req, &resp)
	}

	p.origin.Append(key, value)
}

func (p *sharedDataProxy) GetOwner(key string) string {
	return p.origin.GetOwner(key)
}

func (p *sharedDataProxy) SetMe(me string) {
	panic("Can't be called on proxy")
}

func (p *sharedDataProxy) GetMe() string {
	return p.origin.GetMe()
}

func (p *sharedDataProxy) IsMe(me string) bool {
	return p.origin.IsMe(me)
}

func (p *sharedDataProxy) OnDataChangeCB(cb func(key string)) {
	p.origin.OnDataChangeCB(cb)
}

func (p *sharedDataProxy) ChangeDataOwner(key string, owner string) {
	p.lock.Lock()
	defer p.lock.Unlock()
	// TODO send to others

	req := pb.SharedDataChangeOwner{
		SharedDataId: string(p.origin.GetId()),
		Originator:   p.sdm.GetMe().ToPB(),
		Key:          key,
		Owner:        owner,
	}
	resp := pb.SharedDataChangeOwnerResponse{}
	for _, value := range p.invities {
		p.sdm.clientCache.POST(value.Address, "/shareddata/changeowner", &req, &resp)
	}

	p.origin.ChangeDataOwner(key, owner)
}
