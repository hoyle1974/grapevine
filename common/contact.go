package common

import (
	"fmt"
	"net"

	pb "github.com/hoyle1974/grapevine/proto"
)

type Myself interface {
	GetMe() Contact
}

type Address struct {
	Ip   net.IP
	Port int
}

func (a Address) ToPB() *pb.ClientAddress {
	return &pb.ClientAddress{
		IpAddress: a.Ip.String(),
		Port:      int32(a.Port),
	}
}

func (u Address) GetURL() string {
	return fmt.Sprintf("%s:%d", u.Ip.String(), u.Port)
}

func (s Address) Equal(addr Address) bool {
	return s.Port == addr.Port && s.Ip.Equal(addr.Ip)
}

func (s Address) String() string {
	return fmt.Sprintf("%s:%d", s.Ip.String(), s.Port)
}

type Contact struct {
	AccountId AccountId
	Address   Address
}

func (c Contact) String() string {
	return fmt.Sprintf("%s:%s", c.AccountId, c.Address.String())
}

func NewContact(accountId AccountId, ip net.IP, port int) Contact {
	return Contact{AccountId: accountId, Address: Address{ip, port}}
}

func NewContactFromPB(c *pb.UserContact) Contact {
	return Contact{AccountId: NewAccountId(c.AccountId), Address: NewAddressFromPB(c.ClientAddress)}
}

func NewAddress(ip net.IP, port int) Address {
	return Address{Ip: ip, Port: port}
}

func NewAddressFromPB(c *pb.ClientAddress) Address {
	return NewAddress(net.ParseIP(c.IpAddress), int(c.Port))
}

func (c Contact) ToPB() *pb.UserContact {
	return &pb.UserContact{
		AccountId:     c.AccountId.String(),
		ClientAddress: c.Address.ToPB(),
	}
}

func ContactsToPB(contacts []Contact) []*pb.UserContact {
	out := make([]*pb.UserContact, len(contacts))
	for _, contact := range contacts {
		out = append(out, contact.ToPB())
	}

	return out
}
