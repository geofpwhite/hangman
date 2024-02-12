package hangman

import (
	"fmt"
	"log"
	"net/http"
	"slices"

	// "slices"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// var connections []*websocket.Conn = []*websocket.Conn{}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (gState *gameState) handleWebSocket(
	conn *websocket.Conn,
	inputChannel chan inputInfo,
	timeoutChannel chan int,
	outputChannel chan clientState,
	closeGameChannel chan int,
	removePlayerChannel chan [2]int,
) {
	// first thing to do when we have a new connection is tell them the current game state

	playerIndex := len(gState.players)
	newPlayer := player{username: "Player " + strconv.Itoa(playerIndex+1), connection: conn}
	gState.players = append(gState.players, newPlayer)
	// gState.connections = append(gState.connections, conn)
	if playerIndex == 1 && !gState.needNewWord {
		gState.turn = 1
	}
	defer func() {
		// removePlayerChannel <- [2]int{gState.gameIndex, slices.Index(gState.players, newPlayer)}
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
	}
	for i, player := range gState.players {
		currentState.PlayerIndex = i
		player.connection.WriteJSON(currentState)
	}

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			index := slices.Index(gState.players, newPlayer)
			if index == -1 {

				return
			}
			removePlayerChannel <- [2]int{gState.gameIndex, index}
			return
		}
		i := inputInfo{
			GameIndex: gState.gameIndex,
			PlayerIndex: slices.IndexFunc(gState.players, func(p player) bool {
				return p.connection == conn
			}),
			Username: "",
			Guess:    "",
			Word:     "",
		}
		if err != nil {
			fmt.Println(err)
			return
		}
		switch messageType {
		case websocket.TextMessage:
			pString := string(p)
			if err != nil {
				continue
			}
			switch pString[:2] {
			case "g:":
				i.Guess = pString[2:]
			case "u:":
				i.Username = pString[2:]
			case "w:":
				i.Word = pString[2:]
			default:
				continue
			}

			inputChannel <- i
			// send what the user wants to do to input channel for game function to handle
		}

	}
}

func server(inputChannel chan inputInfo, timeoutChannel chan int, outputChannel chan clientState, newGameChannel chan bool, closeGameChannel chan int, removePlayerChannel chan [2]int) {
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
		fmt.Println("newgame")
		newGameChannel <- true
		c.JSON(200, struct{ length int }{length: len(gStates)})
	})
	r.GET("/get_games", func(c *gin.Context) {
		fmt.Println("games got")
		c.String(http.StatusOK, strconv.Itoa(len(gStates)))
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
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println(err)
			return
		}

		gStates[gameIndex].handleWebSocket(conn, inputChannel, timeoutChannel, outputChannel, closeGameChannel, removePlayerChannel)
		// Handle WebSocket connections here
	})
	go func() {
		for {
			// write changed state to clients
			select {
			case s := <-outputChannel:
				if s.GameIndex > len(gStates) {
					continue
				}
				gState := &gStates[s.GameIndex]

				usernames := []string{}
				for _, p := range (*gState).players {
					usernames = append(usernames, p.username)
				}

				newState := clientState{
					GameIndex:      (*gState).gameIndex,
					Players:        usernames,
					Turn:           (*gState).turn,
					Host:           (*gState).curHostIndex,
					RevealedWord:   (*gState).revealedWord,
					GuessesLeft:    (*gState).guessesLeft,
					LettersGuessed: (*gState).guessed,
					NeedNewWord:    (*gState).needNewWord,
					Warning:        "",
					Winner:         (*gState).winner,
				}
				if newState.NeedNewWord {
					newState.RevealedWord = (*gState).currentWord
				}
				for i, player := range (*gState).players {
					newState.PlayerIndex = i
					if i == s.PlayerIndex {
						newState.Warning = s.Warning
					} else {
						newState.Warning = ""
					}
					if err := player.connection.WriteJSON(newState); err != nil {
						println(err)
					}
				}
			case gameIndex := <-timeoutChannel:
				gState := &gStates[gameIndex]
				usernames := []string{}
				for _, p := range (*gState).players {
					usernames = append(usernames, p.username)
				}
				newState := clientState{
					Players:        usernames,
					Turn:           (*gState).turn,
					Host:           (*gState).curHostIndex,
					RevealedWord:   (*gState).revealedWord,
					GuessesLeft:    (*gState).guessesLeft,
					LettersGuessed: (*gState).guessed,
					NeedNewWord:    (*gState).needNewWord,
					GameIndex:      (*gState).gameIndex,
					Warning:        "timed out",
					Winner:         (*gState).winner,
				}

				for i, player := range (*gState).players {
					newState.PlayerIndex = i
					if err := player.connection.WriteJSON(newState); err != nil {
						println(err)

					}
				}
			}
		}
	}()
	gin.SetMode(gin.ReleaseMode)
	r.Run("localhost:8000") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
