package client

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	protoc "github.com/golang/protobuf/proto"
	"github.com/hoyle1974/grapevine/common"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

type GrapevineClientCache interface {
	GetClient(common.Address) GrapevineClient
	POST(common.Address, string, protoc.Message, protoc.Message) error
}

type grapevineClientCache struct {
	lock    sync.Mutex
	clients map[string]*grapevineClient
}

func NewGrapevineClientCache() GrapevineClientCache {
	return &grapevineClientCache{
		clients: make(map[string]*grapevineClient),
	}
}

type GrapevineClient interface {
	GetClient() *http.Client
}

type grapevineClient struct {
	addr         common.Address
	expiry       time.Time
	roundTripper *http3.RoundTripper
	httpClient   *http.Client
}

func (g *grapevineClient) ResetExpiry() {
}

func (g *grapevineClient) GetClient() *http.Client {
	g.ResetExpiry()
	return g.httpClient
}

func (g *grapevineClientCache) cleanupConnections() {
	for key, value := range g.clients {
		if value.expiry.After(time.Now()) {
			delete(g.clients, key)
			value.roundTripper.Close()
		}
	}
}

func (g *grapevineClientCache) GetClient(addr common.Address) GrapevineClient {
	g.lock.Lock()
	defer g.cleanupConnections()
	defer g.lock.Unlock()

	key := addr.GetURL()

	client, found := g.clients[key]
	if found {
		client.ResetExpiry()
		return client

	}

	client = &grapevineClient{addr: addr, expiry: time.Now().Add(time.Minute)}
	g.clients[key] = client

	pool, err := x509.SystemCertPool()
	if err != nil {
		log.Fatal(err)
	}

	var qconf quic.Config

	client.roundTripper = &http3.RoundTripper{
		TLSClientConfig: &tls.Config{
			RootCAs:            pool,
			InsecureSkipVerify: true,
		},
		QuicConfig: &qconf,
	}

	client.httpClient = &http.Client{
		Transport: client.roundTripper,
	}

	return client
}

// Helper functions to make posts
func (g *grapevineClientCache) POST(addr common.Address, url string, req protoc.Message, gresp protoc.Message) error {
	client := g.GetClient(addr).GetClient()

	b, err := protoc.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := client.Post(fmt.Sprintf("https://%s%s", addr.GetURL(), url), "grpc-message-type", bytes.NewReader(b))
	if err != nil {
		return err
	}

	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = protoc.Unmarshal(b, gresp)
	if err != nil {
		return err
	}

	return nil
}
