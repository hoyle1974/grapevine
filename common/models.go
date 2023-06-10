package common

import "net"

type Contact struct {
	AccountID AccountId
	Ip        net.IP
	Port      int
}
