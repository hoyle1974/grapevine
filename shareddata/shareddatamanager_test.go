package shareddata

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/hoyle1974/grapevine/client"
	"github.com/hoyle1974/grapevine/common"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/stretchr/testify/assert"
)

var globalPorts = 8192
var mutex sync.Mutex

func nextPort() int {
	mutex.Lock()
	defer mutex.Unlock()
	p := globalPorts
	globalPorts++
	return p
}

func newLocalListener(ctx common.CallCtx, sdm SharedDataManager) *http3.Server {
	log := ctx.NewCtx("newLocalListener")
	port := sdm.GetMe().Address.Port
	mux := http.NewServeMux()
	mux.HandleFunc("/shareddata/invite", sdm.OnSharedDataRequestHttp)
	mux.HandleFunc("/shareddata/create", sdm.OnSharedDataRequestHttp)
	mux.HandleFunc("/shareddata/changeowner", sdm.OnSharedDataRequestHttp)
	mux.HandleFunc("/shareddata/set", sdm.OnSharedDataRequestHttp)
	mux.HandleFunc("/shareddata/setmap", sdm.OnSharedDataRequestHttp)
	mux.HandleFunc("/shareddata/append", sdm.OnSharedDataRequestHttp)
	mux.HandleFunc("/shareddata/sendstate", sdm.OnSharedDataRequestHttp)

	quicConf := &quic.Config{
		MaxIdleTimeout: time.Minute * 10,
	}

	addr := fmt.Sprintf("%s:%d", "127.0.0.1", port)

	server := http3.Server{
		Handler:    mux,
		Addr:       addr,
		QuicConfig: quicConf,
	}

	log.Info().Msgf("Listening on %v", server.Addr)

	go func() {
		certFile := "/Users/jstrohm/code/grapevine/grapevine/cert.pem"
		keyFile := "/Users/jstrohm/code/grapevine/grapevine/priv.key"
		err := server.ListenAndServeTLS(certFile, keyFile)
		if err != nil {
			panic(err)
		}
	}()

	// time.Sleep(time.Second)

	return &server
}

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestSharedDataManager(t *testing.T) {
	ctx := common.NewCallCtxWithApp("TestSharedDataManager")

	user1 := common.NewTestMyself("User1", nextPort())
	user2 := common.NewTestMyself("User2", nextPort())

	user1Cb := NewTestClientCb("user1")
	user2Cb := NewTestClientCb("user2")

	cc := client.NewGrapevineClientCache()

	sdmUser1 := NewSharedDataManager(ctx.NewCtx("sdm1"), user1, user1Cb, cc)
	sdmUser2 := NewSharedDataManager(ctx.NewCtx("sdm2"), user2, user2Cb, cc)

	server1 := newLocalListener(ctx, sdmUser1)
	defer server1.CloseGracefully(0)
	server2 := newLocalListener(ctx, sdmUser2)
	defer server2.CloseGracefully(0)

	assert.Equal(t, user1.GetMe(), sdmUser1.GetMe(), "sdmUser1 doesn't match user1!")
	assert.Equal(t, user2.GetMe(), sdmUser2.GetMe(), "sdmUser2 doesn't match user2!")

	sd1 := NewSharedData(user1.GetMe(), "test")
	sd1.Create("key", "bar", "user1", "")
	sdmUser1.Serve(sd1)

	ok := sdmUser1.Invite(sd1, user2.GetMe(), "user2")
	assert.Equal(t, ok, true, "Invite failed")

	assert.Equal(t, sd1.GetId(), user2Cb.sharedDataId, "Shared Data Id did not match")
	assert.Equal(t, "user2", user2Cb.me, "Not the role I expected")

	// time.Sleep(time.Second * 1)

	assert.Equal(t, "bar", sd1.Get("key"), "String didn't match")

	// Try sending some data back and forth
	sd2 := user2Cb.sharedData

	assert.Equal(t, "bar", sd2.Get("key"), "String didn't match")

	sd2.Set("key", "foo")

	// time.Sleep(time.Second * 1)

	assert.Equal(t, "foo", sd1.Get("key"), "String didn't match")
}

type TestUserData struct {
	Name string
}

func TestArrays(t *testing.T) {
	ctx := common.NewCallCtxWithApp("TestArrays")

	// Shared
	cc := client.NewGrapevineClientCache()

	// User 1
	user1 := common.NewTestMyself("User1", nextPort())
	user1Cb := NewTestClientCb("user1")
	sdmUser1 := NewSharedDataManager(ctx.NewCtx("sdm1"), user1, user1Cb, cc)
	server1 := newLocalListener(ctx, sdmUser1)
	defer server1.CloseGracefully(0)

	// User 2
	user2 := common.NewTestMyself("User2", nextPort())
	user2Cb := NewTestClientCb("user2")
	sdmUser2 := NewSharedDataManager(ctx.NewCtx("sdm2"), user2, user2Cb, cc)
	server2 := newLocalListener(ctx, sdmUser2)
	defer server2.CloseGracefully(0)

	// User 1 create the game
	osd1 := NewSharedData(user1.GetMe(), "test")

	gob.Register(TestUserData{})
	gob.Register([]TestUserData{})

	osd1.SetMe("player1")
	osd1.Create("chat", []string{}, "default", "default")
	osd1.Create("users", []TestUserData{}, "default", "default")
	osd1.Create("channel", "bar", "default", "")
	osd1.Create("visibility-group", map[string][]string{"default": []string{"player1", "player2"}}, "system", "default")

	sd1 := sdmUser1.Serve(osd1)

	ok := sdmUser1.Invite(osd1, user2.GetMe(), "player2")
	assert.Equal(t, ok, true, "Invite failed")

	assert.Equal(t, sd1.GetId(), user2Cb.sharedDataId, "Shared Data Id did not match")
	assert.Equal(t, "player2", user2Cb.me, "Not the role I expected")

	sd2 := user2Cb.sharedData

	sd1.Append("chat", "chat from user1")
	sd2.Append("chat", "chat from user2")

	temp := sd1.Get("chat")
	c, ok := temp.([]interface{})
	assert.Equal(t, ok, true, "Chat isn't an array")

	assert.Equal(t, c[0], "chat from user1", "chat 1 is wrong")
	assert.Equal(t, c[1], "chat from user2", "chat 2 is wrong")

	sd1.Append("users", TestUserData{"user1"})
	sd2.Append("users", TestUserData{"user2"})

	temp2 := sd1.Get("users")
	c2, ok := temp2.([]interface{})
	assert.Equal(t, ok, true, "users isn't an array")

	u1, ok := c2[0].(TestUserData)
	assert.Equal(t, ok, true, "users doesn't have User objects")

	u2, ok := c2[1].(TestUserData)
	assert.Equal(t, ok, true, "users doesn't have User objects")

	assert.Equal(t, u1.Name, "user1")
	assert.Equal(t, u2.Name, "user2")

}

func TestMap(t *testing.T) {
	ctx := common.NewCallCtxWithApp("TestMap")

	gob.Register(map[string]interface{}{})

	// Shared
	cc := client.NewGrapevineClientCache()

	// User 1
	user1 := common.NewTestMyself("User1", nextPort())
	user1Cb := NewTestClientCb("user1")
	sdmUser1 := NewSharedDataManager(ctx.NewCtx("sdm1"), user1, user1Cb, cc)
	server1 := newLocalListener(ctx, sdmUser1)
	defer server1.CloseGracefully(0)

	// User 2
	user2 := common.NewTestMyself("User2", nextPort())
	user2Cb := NewTestClientCb("user2")
	sdmUser2 := NewSharedDataManager(ctx.NewCtx("sdm2"), user2, user2Cb, cc)
	server2 := newLocalListener(ctx, sdmUser2)
	defer server2.CloseGracefully(0)

	// User 1 create the game
	osd1 := NewSharedData(user1.GetMe(), "test")

	gob.Register(TestUserData{})
	gob.Register([]TestUserData{})

	osd1.SetMe("player1")
	osd1.CreateMap("map", map[string]string{}, "default", "default")
	osd1.Create("visibility-group", map[string][]string{"default": []string{"player1", "player2"}}, "system", "default")

	sd1 := sdmUser1.Serve(osd1)

	ok := sdmUser1.Invite(osd1, user2.GetMe(), "player2")
	assert.Equal(t, ok, true, "Invite failed")

	assert.Equal(t, sd1.GetId(), user2Cb.sharedDataId, "Shared Data Id did not match")
	assert.Equal(t, "player2", user2Cb.me, "Not the role I expected")

	sd2 := user2Cb.sharedData

	sd1.SetMap("map", "k1", "v1")
	sd2.SetMap("map", "k2", "v2")

	temp := sd1.Get("map")
	c, ok := temp.(map[string]interface{})
	assert.Equal(t, ok, true, "map isn't a map")

	assert.Equal(t, c["k1"], "v1", "k1 is wrong")
	assert.Equal(t, c["k2"], "v2", "k2 is wrong")

}
