package client

import (
	"fmt"

	"github.com/hoyle1974/grapevine/common"
)

type SearchId string
type SharedDataId string

func NewSearchId(id string) SearchId {
	return SearchId(id)
}

func (s SearchId) String() string {
	return string(s)
}

type ClientCallback interface {
	OnSearch(id SearchId, query string) bool
	OnSearchResult(id SearchId, result string, contact common.Contact)
	OnInvited(sharedData SharedData, me string, contact common.Contact) bool
	OnInviteAccepted(sharedData SharedData, contact common.Contact)
}

type SharedData interface {
	IsProxy() bool
	GetCreator() common.Contact
	GetId() SharedDataId
	Create(key string, value interface{}, owner string, visibility string)
	Get(key string) interface{}
	Set(key string, value interface{})
	Append(key string, value interface{})
	GetOwner(key string) string
	SetMe(string)
	GetMe() string
	IsMe(string) bool
	OnDataChangeCB(func(key string))
	ChangeDataOwner(key string, owner string)
	GetData() map[string]data
}

type data struct {
	value      interface{}
	owner      string
	visibility string
}

type sharedData struct {
	creator common.Contact
	id      SharedDataId
	me      string
	data    map[string]data
	cb      func(key string)
}

func NewSharedData(creator common.Contact, id SharedDataId) SharedData {
	return &sharedData{id: id, creator: creator, data: make(map[string]data)}
}

func (s *sharedData) GetData() map[string]data {
	return s.data
}

func (s *sharedData) IsProxy() bool {
	return false
}

func (s *sharedData) GetCreator() common.Contact {
	return s.creator
}

func (s *sharedData) GetId() SharedDataId {
	return s.id
}

func (s *sharedData) Create(key string, value interface{}, owner string, visibility string) {
	fmt.Printf("CREATE %v:%v  (%v:%v)\n", key, value, owner, visibility)
	s.data[key] = data{value, owner, visibility}
}

func (s *sharedData) Set(key string, value interface{}) {
	fmt.Printf("Set %v:%v \n", key, value)
	d, ok := s.data[key]
	if !ok {
		return
	}
	if s.IsMe(d.owner) {
		s.data[key] = data{value, d.owner, d.visibility}
	}
}

func (s *sharedData) Get(key string) interface{} {
	return s.data[key].value
}

func (s *sharedData) Append(key string, value interface{}) {
	fmt.Printf("Append %v:%v \n", key, value)

	d, ok := s.data[key]
	if !ok {
		return
	}
	v := d.value.([]interface{})
	v = append(v, value)
	s.data[key] = data{v, d.owner, d.visibility}
}

func (s *sharedData) GetOwner(key string) string {
	return s.data[key].owner
}

func (s *sharedData) SetMe(me string) {
	s.me = me
}

func (s *sharedData) GetMe() string {
	return s.me
}

func (s *sharedData) IsMe(other string) bool {
	// TODO am I in this group?  is this a group?
	return s.me == other
}

func (s *sharedData) OnDataChangeCB(cb func(string)) {
	s.cb = cb
}

func (s *sharedData) ChangeDataOwner(key string, owner string) {
	fmt.Printf("ChangeDataOwner %v:%v \n", key, owner)

	data, ok := s.data[key]
	if !ok {
		return
	}
	if data.owner != s.me {
		return
	}
	data.owner = owner
}
