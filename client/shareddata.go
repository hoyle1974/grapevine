package client

import "github.com/hoyle1974/grapevine/services"

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
	OnSearchResult(id SearchId, result string, contact services.UserContact)
	OnInvited(sharedData SharedData, me string, contact services.UserContact)
	OnInviteAccepted(sharedData SharedData, contact services.UserContact)
}

type SharedData interface {
	GetCreator() services.AccountId
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
}

type data struct {
	value      interface{}
	owner      string
	visibility string
}

type sharedData struct {
	creator services.AccountId
	id      SharedDataId
	me      string
	data    map[string]data
	cb      func(key string)
}

func NewSharedData(creator services.AccountId) SharedData {
	return &sharedData{creator: creator}
}

func (s *sharedData) GetCreator() services.AccountId {
	return s.creator
}

func (s *sharedData) GetId() SharedDataId {
	return s.id
}

func (s *sharedData) Create(key string, value interface{}, owner string, visibility string) {
	s.data[key] = data{value, owner, visibility}
}

func (s *sharedData) Set(key string, value interface{}) {
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
	if s.me == other {
		return true
	}

	// TODO am I in this group?  is this a group?
	return false
}

func (s *sharedData) OnDataChangeCB(cb func(string)) {
	s.cb = cb
}

func (s *sharedData) ChangeDataOwner(key string, owner string) {
	data, ok := s.data[key]
	if !ok {
		return
	}
	if data.owner != s.me {
		return
	}
	data.owner = owner
}
