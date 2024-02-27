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
	timeoutChannel chan int,
	outputChannel chan clientState,
	closeGameChannel chan int,
	removePlayerChannel chan [2]int,
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
			_player.connection = conn
			playerIndex = slices.IndexFunc(gState.players, func(p player) bool { return p == *_player })
			if playerIndex == -1 {
				conn.WriteJSON(clientState{Hash: ""})
			}
		}

	} else {
		playerIndex = len(gState.players)
		newPlayer := player{username: "Player " + strconv.Itoa(playerIndex+1), connection: conn}
		playerHash := Hash(32)
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
			GameIndex:      gState.gameIndex,
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
		GameIndex:      gState.gameIndex,
		ChatLogs:       gState.chatLogs,
	}

	for i, player := range gState.players {
		currentState.PlayerIndex = i
		player.connection.WriteJSON(currentState)
	}

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			index := slices.IndexFunc(gState.players, func(p player) bool { return p.connection == conn })
			if index == -1 {
				return
			}
			// removePlayerChannel <- [2]int{gState.gameIndex, index}

			return
		}
		i := inputInfo{
			GameIndex: gState.gameIndex,
			PlayerIndex: slices.IndexFunc(gState.players, func(p player) bool {
				return p.connection == conn
			}),
		}
		switch messageType {
		case websocket.TextMessage:
			pString := string(p)
			switch pString[:2] {
			case "g:":
				i.Guess = pString[2:]
				inp := guessInput{GameIndex: i.GameIndex, PlayerIndex: i.PlayerIndex, Guess: i.Guess}
				inputChannel <- &inp
			case "u:":
				i.Username = pString[2:]
				inp := usernameInput{GameIndex: i.GameIndex, PlayerIndex: i.PlayerIndex, Username: i.Username}
				inputChannel <- &inp
			case "w:":
				i.Word = pString[2:]
				inp := newWordInput{GameIndex: i.GameIndex, PlayerIndex: i.PlayerIndex, NewWord: i.Word}
				inputChannel <- &inp
			case "c:":
				i.Chat = pString[2:]
				inp := chatInput{GameIndex: i.GameIndex, PlayerIndex: i.PlayerIndex, Message: i.Chat}
				inputChannel <- &inp
			default:
				continue
			}
		}
	}
}

func server(inputChannel chan input, timeoutChannel chan int, outputChannel chan clientState, newGameChannel chan bool, closeGameChannel chan int, removePlayerChannel chan [2]int) {
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
		newGameChannel <- true
		c.JSON(200, struct{ length int }{length: len(gStates)})
	})
	r.GET("/get_games", func(c *gin.Context) {
		c.String(http.StatusOK, strconv.Itoa(len(gStates)))
	})

	r.GET("/reconnect/:playerHash", func(c *gin.Context) {
		playerHash, b := c.Params.Get("playerHash")
		if !b {
			return
		}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Println(err)
			conn.Close()
			return
		}
		gameIndex := -1
		for i := range gStates {
			if index := slices.IndexFunc(gStates[i].players, func(p player) bool { return p == *hashes[playerHash] }); index != -1 {
				gameIndex = i
				break
			}
		}
		if gameIndex >= 0 {
			handleWebSocket(conn, inputChannel, timeoutChannel, outputChannel, closeGameChannel, removePlayerChannel, gStates[gameIndex], true, playerHash)
		} else {
			c.String(http.StatusOK, "failed")
		}

	})

	r.GET("/valid/:playerHash", func(c *gin.Context) {
		hash, _ := c.Params.Get("playerHash")
		if hashes[hash] == nil {
			c.String(http.StatusOK, "-1")
		} else {
			var gameIndex int
			for i, g := range gStates {
				for _, p := range g.players {
					if p == *hashes[hash] {
						gameIndex = i
					}
				}
			}
			c.String(http.StatusOK, strconv.Itoa(gameIndex))
		}
	})

	r.GET("/exit_game/:playerHash/:gameIndex", func(c *gin.Context) {
		defer c.String(http.StatusOK, "ok")
		playerHash, _ := c.Params.Get("playerHash")
		str, _ := c.Params.Get("gameIndex")
		gameIndex, err := strconv.Atoi(str)
		if err != nil {
			panic("ahhhhh")
		}
		_player := hashes[playerHash]
		if _player == nil {
			return
		}
		playerIndex := slices.IndexFunc(gStates[gameIndex].players, func(p player) bool { return p == *_player })
		delete(hashes, playerHash)
		removePlayerChannel <- [2]int{gameIndex, playerIndex}
	})
	r.GET("/ws/:gameIndex", func(c *gin.Context) {
		str, b := c.Params.Get("gameIndex")
		if !b {
			return
		}
		gameIndex, err := strconv.Atoi(str)
		if err != nil {
			return
		}
		if gameIndex >= len(gStates) {
			c.String(http.StatusOK, strconv.Itoa(len(gStates)))
			return
		}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Println(err)
			conn.Close()
			return
		}
		handleWebSocket(conn, inputChannel, timeoutChannel, outputChannel, closeGameChannel, removePlayerChannel, gStates[gameIndex], false, "")
		// Handle WebSocket connections here
	})

	go outputLoop(timeoutChannel, outputChannel)
	gin.SetMode(gin.ReleaseMode)
	r.Run("localhost:8080") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

	// r.RunTLS("localhost:8080") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func outputLoop(
	timeoutChannel chan int,
	outputChannel chan clientState,
) {
	for {
		// write changed state to clients
		select {
		case s := <-outputChannel:
			log.Println("outputChannel")
			if s.GameIndex >= len(gStates) {
				continue
			}
			gState := gStates[s.GameIndex]

			usernames := []string{}
			for _, p := range gState.players {
				usernames = append(usernames, p.username)
			}
			newState := clientState{
				GameIndex:      gState.gameIndex,
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
				}
			}
		case gameIndex := <-timeoutChannel:
			log.Println("timeoutChannel")
			if gameIndex >= len(gStates) || gameIndex < 0 {
				continue
			}
			gState := gStates[gameIndex]
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
				GameIndex:      gState.gameIndex,
				Warning:        "timed out",
				Winner:         gState.winner,
				ChatLogs:       gState.chatLogs,
			}

			gState.mut.Lock()
			for i, player := range (*gState).players {
				newState.PlayerIndex = i
				if err := player.connection.WriteJSON(newState); err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}
