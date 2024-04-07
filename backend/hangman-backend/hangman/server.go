package hangman

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var hashes map[string]*player = map[string]*player{}

/*
Create random hash string for user
*/
func Hash(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func handleWebSocket(
	conn *websocket.Conn,
	inputChannel chan input,
	gState *gameState,
	reconnect bool,
	hash string,
) {
	// first thing to do when we have a new connection is tell them the current game state

	log.Println("reconnect is ", reconnect)
	var playerIndex int
	if reconnect {
		_player := hashes[hash]
		if _player != nil {
			if _player.connection != nil {
				if err := _player.connection.Close(); err != nil {
					fmt.Println(err)
				}
			}
			playerIndex = slices.IndexFunc(gState.players, func(p player) bool { return p.hash == _player.hash })
			_player.connection = conn
			hashes[hash] = _player
			if playerIndex == -1 {
				conn.WriteJSON(clientState{Hash: "undefined", Warning: "2"})
				fmt.Println(_player)
				conn.Close()
				return
			}
			gState.players[playerIndex].connection = conn
		}

	} else {
		playerIndex = len(gState.players)
		playerHash := Hash(32)
		hash = playerHash
		newPlayer := player{username: "Player " + strconv.Itoa(playerIndex+1), connection: conn, hash: playerHash}
		gState.newPlayer(newPlayer)

		hashes[playerHash] = &gState.players[playerIndex]
		usernames := []string{}
		for _, p := range gState.players {
			usernames = append(usernames, p.username)
		}
		newPlayer.connection.WriteJSON(clientState{
			Players:        usernames,
			Turn:           gState.turn,
			Host:           gState.curHostIndex,
			RevealedWord:   gState.revealedWord,
			GuessesLeft:    gState.guessesLeft,
			LettersGuessed: gState.guessed,
			NeedNewWord:    gState.needNewWord,
			Warning:        "",
			PlayerIndex:    playerIndex,
			Winner:         gState.winner,
			GameHash:       gState.gameHash,
			ChatLogs:       gState.chatLogs,
			Hash:           playerHash,
		})

	}
	// gState.connections = append(gState.connections, conn)
	defer func() {
		conn.Close()
	}()
	usernames := []string{}
	for _, p := range gState.players {
		usernames = append(usernames, p.username)
	}

	currentState := clientState{
		Players:        usernames,
		Turn:           gState.turn,
		Host:           gState.curHostIndex,
		RevealedWord:   gState.revealedWord,
		GuessesLeft:    gState.guessesLeft,
		LettersGuessed: gState.guessed,
		NeedNewWord:    gState.needNewWord,
		Warning:        "",
		PlayerIndex:    playerIndex,
		Winner:         gState.winner,
		GameHash:       gState.gameHash,
		ChatLogs:       gState.chatLogs,
	}

	for i, player := range gState.players {
		currentState.PlayerIndex = i
		player.connection.WriteJSON(currentState)
	}

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			return
		}

		i := inputInfo{
			GameHash: gState.gameHash,
			PlayerIndex: slices.IndexFunc(gState.players, func(p player) bool {
				return p.hash == hash
			}),
		}
		switch messageType {
		case websocket.TextMessage:
			pString := string(p)
			switch pString[:2] {
			case "g:":
				i.Guess = pString[2:]
				inp := guessInput{GameHash: i.GameHash, PlayerIndex: i.PlayerIndex, Guess: i.Guess}
				inputChannel <- &inp
			case "u:":
				i.Username = pString[2:]
				inp := usernameInput{GameHash: i.GameHash, PlayerIndex: i.PlayerIndex, Username: i.Username}
				inputChannel <- &inp
			case "w:":
				i.Word = pString[2:]
				inp := newWordInput{GameHash: i.GameHash, PlayerIndex: i.PlayerIndex, NewWord: i.Word}
				inputChannel <- &inp
			case "c:":
				i.Chat = pString[2:]
				inp := chatInput{GameHash: i.GameHash, PlayerIndex: i.PlayerIndex, Message: i.Chat}
				inputChannel <- &inp
			case "r:":
				inp := randomlyChooseWordInput{GameHash: i.GameHash, PlayerIndex: i.PlayerIndex}
				inputChannel <- &inp

			default:
				continue
			}
		}
	}
}

func server(inputChannel chan input, timeoutChannel chan string, outputChannel chan clientState, newGameChannel chan bool, removePlayerChannel chan playerToRemove, tickerTimeoutChannel chan string, closeGameChannel chan string, tickerInputChannels map[string]chan inputInfo) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
	})
	r.GET("/new_game", func(c *gin.Context) {
		log.Println("newGameChannel")
		gState := newGame()
		newTickerInputChannel := make(chan (inputInfo))
		tickerInputChannels[gState.gameHash] = newTickerInputChannel
		go (*gState).runTicker(tickerTimeoutChannel, newTickerInputChannel, closeGameChannel)
		c.JSON(200, struct{ gameHash string }{gameHash: gState.gameHash})
	})
	r.GET("/get_games", func(c *gin.Context) {
		c.String(http.StatusOK, "0")
	})

	r.GET("/reconnect/:playerHash/:gameHash", func(c *gin.Context) {

		playerHash, b := c.Params.Get("playerHash")
		if !b {
			return
		}
		gameHash, b := c.Params.Get("gameHash")
		if !b {
			return
		}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Println(err)
			fmt.Println(2)
			conn.Close()
			return
		}

		if index := slices.IndexFunc(
			gameHashes[gameHash].players,
			func(p player) bool {
				fmt.Println(p, hashes[playerHash], playerHash)
				return p.hash == playerHash
			}); index != -1 {

			conn.WriteJSON(clientState{Hash: "undefined", Warning: "1"})
			return
		}

		if gameHashes[gameHash] != nil {
			handleWebSocket(conn, inputChannel, gameHashes[gameHash], true, playerHash)
		} else {
			conn.WriteJSON(clientState{Hash: "undefined", Warning: "1"})
		}
	})

	r.GET("/valid/:playerHash", func(c *gin.Context) {
		hash, _ := c.Params.Get("playerHash")
		if hashes[hash] == nil {
			c.String(http.StatusOK, "-1")
		} else {
			var gameHash string
			for i, g := range gameHashes {
				for _, p := range g.players {
					if p == *hashes[hash] {
						gameHash = i
					}
				}
			}
			c.String(http.StatusOK, gameHash)
		}
	})

	r.GET("/exit_game/:playerHash/:gameHash", func(c *gin.Context) {
		defer c.String(http.StatusOK, "ok")
		playerHash, _ := c.Params.Get("playerHash")
		gameHash, _ := c.Params.Get("gameHash")
		_player := hashes[playerHash]
		if _player == nil || gameHashes[gameHash] == nil {
			return
		}
		playerIndex := slices.IndexFunc(gameHashes[gameHash].players, func(p player) bool { return p.hash == _player.hash })
		delete(hashes, playerHash)
		removePlayerChannel <- playerToRemove{gameHash, playerIndex}
	})
	r.GET("/ws/:gameHash", func(c *gin.Context) {
		gameHash, b := c.Params.Get("gameHash")
		if !b {
			return
		}
		if gameHashes[gameHash] == nil {
			c.String(http.StatusOK, "no such game")
			return
		}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Println(err)
			fmt.Println(4)
			conn.Close()
			return
		}
		handleWebSocket(conn, inputChannel, gameHashes[gameHash], false, "")
		// Handle WebSocket connections here
	})

	go outputLoop(timeoutChannel, outputChannel)
	gin.SetMode(gin.ReleaseMode)
	r.Run("localhost:8080") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	// r.RunTLS("localhost:8080") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func outputLoop(
	timeoutChannel chan string,
	outputChannel chan clientState,
) {
	for {
		// write changed state to clients
		select {
		case s := <-outputChannel:
			log.Println("outputChannel")
			gState := gameHashes[s.GameHash]
			if gState == nil {
				continue
			}

			usernames := []string{}
			for _, p := range gState.players {
				usernames = append(usernames, p.username)
			}
			newState := clientState{
				GameHash:       gState.gameHash,
				Players:        usernames,
				Turn:           gState.turn,
				Host:           gState.curHostIndex,
				RevealedWord:   gState.revealedWord,
				GuessesLeft:    gState.guessesLeft,
				LettersGuessed: gState.guessed,
				NeedNewWord:    gState.needNewWord,
				Warning:        "",
				Winner:         gState.winner,
				ChatLogs:       gState.chatLogs,
			}
			if newState.NeedNewWord {
				newState.RevealedWord = gState.currentWord
			}
			for i, player := range gState.players {
				newState.PlayerIndex = i
				if i == s.PlayerIndex {
					newState.Warning = s.Warning
				} else {
					newState.Warning = ""
				}
				if err := player.connection.WriteJSON(newState); err != nil {
					fmt.Println(err)
					fmt.Println(3)
				}
			}
		case gameHash := <-timeoutChannel:
			log.Println("timeoutChannel")
			if gameHashes[gameHash] == nil {
				continue
			}
			gState := gameHashes[gameHash]
			usernames := []string{}
			for _, p := range (*gState).players {
				usernames = append(usernames, p.username)
			}
			newState := clientState{
				Players:        usernames,
				Turn:           gState.turn,
				Host:           gState.curHostIndex,
				RevealedWord:   gState.revealedWord,
				GuessesLeft:    gState.guessesLeft,
				LettersGuessed: gState.guessed,
				NeedNewWord:    gState.needNewWord,
				GameHash:       gState.gameHash,
				Warning:        "timed out",
				Winner:         gState.winner,
				ChatLogs:       gState.chatLogs,
			}

			gState.mut.Lock()
			for i, player := range (*gState).players {
				newState.PlayerIndex = i
				if err := player.connection.WriteJSON(newState); err != nil {
					fmt.Println(err)
					fmt.Println(4)
				}
			}
		}
	}
}
