package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"runtime/debug"

	"github.com/hoyle1974/grapevine/client"
	"github.com/hoyle1974/grapevine/common"
	"github.com/rs/zerolog"
)

func GetOutboundIP(ctx client.CallCtx) net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		if err.Error() == "dial udp 8.8.8.8:80: connect: network is unreachable" {
			return net.ParseIP("127.0.0.1")
		}
		ctx.Fatal().Err(err)
		panic(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

type Callback struct {
	lock       sync.Mutex
	searching  bool
	sharedData client.SharedData
	grapevine  client.Grapevine
}

const gameType = "grapevine.com/game/example/tictactoe/v1"

// Someone is searching for this query
func (c *Callback) OnSearch(id client.SearchId, query string) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	if !c.searching {
		fmt.Println("We are done searching!")
		return false // We are done searching
	}
	if query == gameType {
		fmt.Println("We support this game!")
		return true // We support this game type
	}
	fmt.Printf("We do not support %v!\n", query)
	return false // Not for us
}

// We found someone matching our game type search
func (c *Callback) OnSearchResult(id client.SearchId, query string, contact common.Contact) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if !c.searching {
		return
	}
	c.searching = false
	fmt.Printf("Id: %v Query: %v\n", id, query)

	// Let's try starting a game with this client and see if they will accept our invitation
	c.sharedData = client.NewSharedData(c.sharedData.GetCreator()) // Init the structure
	c.sharedData.SetMe("player1")
	c.sharedData.Create("state", "start", c.sharedData.GetMe(), "default")
	c.sharedData.Create("board", ".........", c.sharedData.GetMe(), "default")
	c.sharedData.Create("chat", []string{}, "default", "default")
	c.sharedData.Create("visibility-group", map[string][]string{"default": []string{"player1", "player2"}}, "system", "default")
	c.sharedData = c.grapevine.Serve(c.sharedData) // The structure is now live and can be worked with by this client or others

	// Invite your contact to join the structure as player2
	if !c.grapevine.Invite(c.sharedData, contact, "player2") {
		c.grapevine.LeaveShare(c.sharedData)
	}
}

// Someone has invited us to share data (in our case it's a game
func (c *Callback) OnInvited(sharedData client.SharedData, me string, contact common.Contact) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.sharedData != nil {
		// Leave any existing share we have, probably not the best choice
		c.grapevine.LeaveShare(c.sharedData)
		c.sharedData = nil
	}

	// Agree to join with this shared data
	c.grapevine.JoinShare(c.sharedData)

	c.sharedData = sharedData

	c.play()
}

// Someone accepted our invitation to share the data
func (c *Callback) OnInviteAccepted(sharedData client.SharedData, contact common.Contact) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if sharedData.GetId() != c.sharedData.GetId() {
		fmt.Printf("Unexpected invite received: %v vs %v\n", sharedData.GetId(), c.sharedData.GetId())
		return
	}

	// Let's start the game
	fmt.Printf("Player %v has joined", contact.AccountId)
	c.sharedData.Set("state", "player1")

	c.play()
}

func getInput() (string, string) {
	fmt.Printf("Input>")
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')

	// convert CRLF to LF
	text = strings.Replace(text, "\n", "", -1)

	temp := strings.SplitN(text, " ", 2)

	return temp[0], temp[1]
}

func doMove(move string, board string, piece string) string {
	idx, err := strconv.Atoi(move)
	if err != nil {
		fmt.Println(err)
		return board
	}
	a := ""
	b := piece
	c := ""
	if idx > 0 {
		a = board[:idx-1]
	}
	c = board[idx+1:]
	return a + b + c
}

func myPiece(player string) string {
	if player == "player1" {
		return "X"
	}
	return "O"
}

func didIWin(board string, piece string) bool {
	// 0 1 2
	// 3 4 5
	// 6 7 8

	if board[0] == piece[0] {
		if board[0] == board[1] && board[1] == board[2] {
			return true
		}
		if board[0] == board[4] && board[4] == board[8] {
			return true
		}
		if board[0] == board[3] && board[3] == board[6] {
			return true
		}
	}
	if board[1] == piece[0] {
		if board[1] == board[4] && board[4] == board[7] {
			return true
		}
	}
	if board[3] == piece[0] {
		if board[3] == board[4] && board[4] == board[5] {
			return true
		}
	}

	return false
}

func (c *Callback) play() {
	c.sharedData.OnDataChangeCB(func(key string) {
		if key == "chat" {
			chat := c.sharedData.Get("chat").([]string)
			fmt.Println("Chat: " + chat[len(chat)-1])
		}
	})

	for {
		// Wait for our turn to play
		for {
			// Wait till we own the state (this means it's either or turn or the game is over and all game objects are owned by everyone
			time.Sleep(time.Second)
			if c.sharedData.IsMe(c.sharedData.GetOwner("state")) {
				break
			}
		}
		if c.sharedData.Get("state").(string) == c.sharedData.GetMe() {
			fmt.Println("Your turn, make a move or chat:")
		} else if c.sharedData.Get("state") == "finished" {
			fmt.Println("Game is over, you may still chat:")
		}

		// Take a turn, blocking on input
		input, extra := getInput()
		if input == "chat" {
			// Append to the chat array
			c.sharedData.Append("chat", extra)
		} else if input == "move" {
			if c.sharedData.Get("state").(string) != c.sharedData.GetMe() {
				fmt.Println("Not our turn to move")
				continue
			}

			// Make a move
			piece := myPiece(c.sharedData.GetMe())
			c.sharedData.Set("board", doMove(extra, c.sharedData.Get("board").(string), piece))
			if didIWin(c.sharedData.Get("board").(string), piece) {
				c.sharedData.Set("state", "finished")
				c.sharedData.ChangeDataOwner("board", "default")
				c.sharedData.ChangeDataOwner("state", "default")
				fmt.Println("You won!")
			} else {
				// Other player can move
				c.sharedData.Set("state", "player2")
			}

			c.sharedData.ChangeDataOwner("board", "player2")
			c.sharedData.ChangeDataOwner("state", "player2")
		} else if input == "leave" {
			c.grapevine.LeaveShare(c.sharedData)
		}
	}
}

func main() {
	flag.Parse()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	ctx := client.NewCallCtxWithApp("tictactoe")

	ctx.Info().Msg("Flags:")
	flag.CommandLine.VisitAll(func(flag *flag.Flag) {
		ctx.Info().Msg(fmt.Sprintf("\t%v:%v", flag.Name, flag.Value))
	})

	info, ok := debug.ReadBuildInfo()
	if !ok {
		panic("couldn't read build info")
	}
	ctx.Info().Msg("Build Info Version: " + info.Main.Version + " " + info.Main.Sum)

	cb := &Callback{searching: true}
	cb.grapevine = client.NewGrapevine(cb, ctx)
	ip := GetOutboundIP(ctx)
	ctx.Info().Msgf("Outbound IP is: %v", ip)
	port, err := cb.grapevine.Start(ip)
	if err != nil {
		ctx.Error().Err(err).Msg("Error starting grapevine")
	}

	username := fmt.Sprintf("U%d", rand.Int()%1000000)
	password := "P" + uuid.New().String()
	ctx.Info().Msgf("Creating account with User %v and Password %v", username, password)
	err = cb.grapevine.CreateAccount(username, password)
	if err != nil {
		ctx.Error().Err(err).Msg("Error creating account")
		return
	}

	ctx.Info().Msg("Logging in")
	accountId, err := cb.grapevine.Login(username, password, ip, port)
	if err != nil {
		ctx.Error().Err(err).Msg("Error logging in")
		return
	}
	ctx.Info().Msgf("Logged in to account: %s", accountId.String())

	cb.grapevine.Search(gameType)

	for {
		time.Sleep(time.Minute)
	}
}
