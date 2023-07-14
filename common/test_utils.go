package common

import "net"

type TestMyself struct {
	me Contact
}

func (m *TestMyself) GetMe() Contact {
	return m.me
}

func NewTestMyself(id string, port int) Myself {
	return &TestMyself{
		me: NewContact(NewAccountId(id), net.ParseIP("127.0.0.1"), port),
	}
}
