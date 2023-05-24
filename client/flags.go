package client

import "flag"

// const ACCOUNT_URL = "account.default.svc.cluster.local:8080"
// const AUTH_URL = "auth.default.svc.cluster.local:8080"
// const GOSSIP_ADDR = "tictactoe.default.svc.cluster.local"

//const ACCOUNT_URL = "localhost:8081"
//const AUTH_URL = "localhost:8080"
//const GOSSIP_ADDR = "localhost"

var (
	accountURL = flag.String("account_url", "localhost:8081", "The address of the account service")
	authURL    = flag.String("auth_url", "localhost:8080", "The address of the auth service")
	gossipAddr = flag.String("gossip_addr", "localhost", "address to try to bootstrap gossip")
)
