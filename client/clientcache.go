package client

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/hoyle1974/grapevine/services"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

type GrapevineClientCache interface {
	GetClient(services.UserContact) GrapevineClient
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
	contact      services.UserContact
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

func (g *grapevineClientCache) GetClient(contact services.UserContact) GrapevineClient {
	g.lock.Lock()
	defer g.cleanupConnections()
	defer g.lock.Unlock()

	key := contact.GetURL()

	client, found := g.clients[key]
	if found {
		client.ResetExpiry()
		return client

	}

	client = &grapevineClient{contact: contact, expiry: time.Now().Add(time.Minute)}
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
