package grapevine

import "flag"

var (
	accountURL = flag.String("account_url", "localhost:8081", "The address of the account service")
	authURL    = flag.String("auth_url", "localhost:8080", "The address of the auth service")
	gossipAddr = flag.String("gossip_addr", "localhost", "address to try to bootstrap gossip")
)
