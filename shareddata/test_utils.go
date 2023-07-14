package shareddata

import (
	"github.com/hoyle1974/grapevine/common"
)

type TestClientCallback struct {
	name         string
	sharedData   SharedData
	sharedDataId SharedDataId
	me           string
}

func (cb *TestClientCallback) OnSearch(id SearchId, query string) bool {
	return true
}
func (cb *TestClientCallback) OnSearchResult(id SearchId, result string, contact common.Contact) {

}
func (cb *TestClientCallback) OnInvited(sharedDataId SharedDataId, me string, contact common.Contact) bool {
	// fmt.Println("--------- OnInvited: " + cb.name)
	cb.sharedDataId = sharedDataId
	cb.me = me

	return true
}
func (cb *TestClientCallback) OnInviteAccepted(sharedData SharedData, contact common.Contact) {
	// fmt.Println("--------- OnInviteAccepted: " + cb.name)
}
func (cb *TestClientCallback) OnSharedDataAvailable(sharedData SharedData) {
	cb.sharedData = sharedData
}

func NewTestClientCb(name string) *TestClientCallback {
	return &TestClientCallback{name: name}
}
