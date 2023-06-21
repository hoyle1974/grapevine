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

	"github.com/hoyle1974/grapevine/common"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"google.golang.org/protobuf/proto"
)

type GrapevineClientCache interface {
	GetClient(common.Address) GrapevineClient
	POST(common.Address, string, proto.Message, proto.Message) error
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
func (g *grapevineClientCache) POST(addr common.Address, url string, req proto.Message, gresp proto.Message) error {
	// fmt.Printf("*** POST %s\n", fmt.Sprintf("https://%s%s", addr.GetURL(), url))
	client := g.GetClient(addr).GetClient()

	b, err := proto.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := client.Post(fmt.Sprintf("https://%s%s", addr.GetURL(), url), "grpc-message-type", bytes.NewReader(b))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}

	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if gresp == nil {
		return nil // We don't care about the response
	}
	err = proto.Unmarshal(b, gresp)
	if err != nil {
		return err
	}

	return nil
}
