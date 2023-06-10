package common

import (
	"fmt"
	"net"

	pb "github.com/hoyle1974/grapevine/proto"
)

type Address struct {
	Ip   net.IP
	Port int
}

type Contact struct {
	AccountId AccountId
	Address   Address
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

func NewContact(accountId AccountId, ip net.IP, port int) Contact {
	return Contact{AccountId: accountId, Address: Address{ip, port}}
}

func NewAddress(ip net.IP, port int) Address {
	return Address{Ip: ip, Port: port}
}

func (c Contact) ToPB() *pb.UserContact {
	return &pb.UserContact{
		UserId: c.AccountId.String(),
		ClientAddress: &pb.ClientAddress{
			IpAddress: c.Address.Ip.String(),
			Port:      int32(c.Address.Port),
		},
	}
}

func ContactsToPB(contacts []Contact) []*pb.UserContact {
	out := make([]*pb.UserContact, len(contacts))
	for _, contact := range contacts {
		out = append(out, contact.ToPB())
	}

	return out
}
