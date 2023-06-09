package shareddata

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sync"

	"github.com/hoyle1974/grapevine/common"
	pb "github.com/hoyle1974/grapevine/proto"
)

type ValueHolder struct {
	Value interface{}
}

func init() {
	gob.Register(ValueHolder{})
	gob.Register(map[string][]string{})
}

func toBytes(value interface{}) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(ValueHolder{value}); err != nil {
		panic(fmt.Sprintf("toBytes: (%v) : %v", value, err))
	}
	return buf.Bytes()
}

func fromBytes(value []byte) interface{} {
	buf := bytes.NewBuffer(value)
	dec := gob.NewDecoder(buf)

	var out ValueHolder
	if err := dec.Decode(&out); err != nil {
		panic(fmt.Sprintf("fromBytes: (%v) : %v", value, err))
	}

	return out.Value
}

type SharedDataProxy interface {
	SharedData
	GetOrigin() SharedData
	AddInvitee(recipient common.Contact, as string)
	SendStateTo(recipient common.Contact)
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

func (p *sharedDataProxy) SendStateTo(recipient common.Contact) {
	p.lock.Lock()
	defer p.lock.Unlock()

	req := pb.SharedDataSendState{
		SharedDataId: string(p.origin.GetId()),
		Originator:   p.sdm.GetMe().ToPB(),
		Data:         make(map[string]*pb.SharedDataData),
		Listeners:    make(map[string]*pb.UserContact),
	}

	for key, value := range p.GetOrigin().GetData() {
		data := &pb.SharedDataData{
			Value:     toBytes(value.value),
			Owner:     value.owner,
			Visbility: value.visibility,
		}
		req.Data[key] = data
	}

	for key, value := range p.invities {
		contact := value.ToPB()
		req.Listeners[key] = contact
	}

	resp := pb.SharedDataSendStateResponse{}
	err := p.sdm.clientCache.POST(recipient.Address, "/shareddata/sendstate", &req, &resp)
	if err != nil {
		panic(err)
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

func (p *sharedDataProxy) GetOrigin() SharedData {
	return p.origin
}

func (s *sharedDataProxy) GetData() map[string]data {
	return s.origin.GetData()
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
		Value:        toBytes(value),
		Owner:        owner,
		Visibility:   visibility,
	}
	resp := pb.SharedDataCreateResponse{}
	for key, value := range p.invities {
		if key != p.GetMe() {
			p.sdm.clientCache.POST(value.Address, "/shareddata/create", &req, &resp)
		}
	}
}

func (p *sharedDataProxy) CreateArray(key string, value []interface{}, owner string, visibility string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.origin.CreateArray(key, value, owner, visibility)

	req := pb.SharedDataCreateArray{
		SharedDataId: string(p.origin.GetId()),
		Originator:   p.sdm.GetMe().ToPB(),
		Key:          key,
		Value:        toBytes(value),
		Owner:        owner,
		Visibility:   visibility,
	}
	resp := pb.SharedDataCreateArrayResponse{}
	for key, value := range p.invities {
		if key != p.GetMe() {
			p.sdm.clientCache.POST(value.Address, "/shareddata/create", &req, &resp)
		}
	}
}

func (p *sharedDataProxy) CreateMap(key string, value interface{}, owner string, visibility string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.origin.CreateMap(key, value, owner, visibility)

	req := pb.SharedDataCreateMap{
		SharedDataId: string(p.origin.GetId()),
		Originator:   p.sdm.GetMe().ToPB(),
		Key:          key,
		Value:        toBytes(value),
		Owner:        owner,
		Visibility:   visibility,
	}
	resp := pb.SharedDataCreateMapResponse{}
	for key, value := range p.invities {
		if key != p.GetMe() {
			p.sdm.clientCache.POST(value.Address, "/shareddata/create", &req, &resp)
		}
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
		Value:        toBytes(value),
	}
	resp := pb.SharedDataSetResponse{}
	for key, value := range p.invities {
		if key != p.GetMe() {
			p.sdm.clientCache.POST(value.Address, "/shareddata/set", &req, &resp)
		}
	}
}

func (p *sharedDataProxy) SetMap(key string, mapKey string, value interface{}) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.origin.SetMap(key, mapKey, value)

	req := pb.SharedDataSetMap{
		SharedDataId: string(p.origin.GetId()),
		Originator:   p.sdm.GetMe().ToPB(),
		Key:          key,
		MapKey:       mapKey,
		Value:        toBytes(value),
	}
	resp := pb.SharedDataSetMapResponse{}
	for key, value := range p.invities {
		if key != p.GetMe() {
			p.sdm.clientCache.POST(value.Address, "/shareddata/setmap", &req, &resp)
		}
	}
}

func (p *sharedDataProxy) Append(key string, value interface{}) {
	p.lock.Lock()
	defer p.lock.Unlock()

	req := pb.SharedDataAppend{
		SharedDataId: string(p.origin.GetId()),
		Originator:   p.sdm.GetMe().ToPB(),
		Key:          key,
		Value:        toBytes(value),
	}
	resp := pb.SharedDataAppendResponse{}
	for key, value := range p.invities {
		if key != p.GetMe() {
			p.sdm.clientCache.POST(value.Address, "/shareddata/append", &req, &resp)
		}
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
	for key, value := range p.invities {
		if key != p.GetMe() {
			p.sdm.clientCache.POST(value.Address, "/shareddata/changeowner", &req, &resp)
		}
	}

	p.origin.ChangeDataOwner(key, owner)
}
