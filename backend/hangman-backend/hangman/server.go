package hangman

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var connections []*websocket.Conn = []*websocket.Conn{}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleWebSocket(conn *websocket.Conn, inputChannel chan info, timeoutChannel chan bool, outputChannel chan clientState) {
	// first thing to do when we have a new connection is tell them the current game state

	connections = append(connections, conn)
	playerIndex := len(sState.players)
	sState.players = append(sState.players, "Player "+strconv.Itoa(playerIndex+1))
	defer func() {
		if len(sState.players)-1 == playerIndex {
			sState.players, connections = sState.players[:playerIndex], connections[:playerIndex]
		} else {
			sState.players = append(sState.players[:playerIndex], sState.players[playerIndex+1:]...)
			connections = append(connections[:playerIndex], connections[playerIndex+1:]...)
		}
		outputChannel <- clientState{}
	}()

	currentState := clientState{
		Players:        sState.players,
		Turn:           sState.turn,
		Host:           sState.curHostIndex,
		RevealedWord:   sState.revealedWord,
		GuessesLeft:    sState.guessesLeft,
		LettersGuessed: sState.guessed,
		NeedNewWord:    sState.needNewWord,
		Warning:        "",
		PlayerIndex:    playerIndex,
	}
	for i, c := range connections {
		currentState.PlayerIndex = i
		c.WriteJSON(currentState)
	}

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			print("er")
		}
		i := info{
			PlayerIndex: playerIndex,
			Username:    "",
			Guess:       "",
			Word:        "",
		}
		fmt.Println(string(p))
		if err != nil {
			fmt.Println(err)
			return
		}
		switch messageType {
		case websocket.TextMessage:
			switch string(p)[:2] {
			case "g:":
				i.Guess = string(p[2:])
			case "u:":
				i.Username = string(p)[2:]
			case "w:":
				i.Word = string(p)[2:]
			default:
				continue
			}

			inputChannel <- i
			// send what the user wants to do to input channel for game function to handle
		}

	}
}

func server(inputChannel chan info, timeoutChannel chan bool, outputChannel chan clientState) {
	r := gin.Default()
	r.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()

		handleWebSocket(conn, inputChannel, timeoutChannel, outputChannel)
		// Handle WebSocket connections here
	})
	go func() {
		for {
			// write changed state to clients
			select {
			case s := <-outputChannel:
				newState := clientState{
					Players:        sState.players,
					Turn:           sState.turn,
					Host:           sState.curHostIndex,
					RevealedWord:   sState.revealedWord,
					GuessesLeft:    sState.guessesLeft,
					LettersGuessed: sState.guessed,
					NeedNewWord:    sState.needNewWord,
					Warning:        "",
					Winner:         sState.winner,
				}
				if newState.NeedNewWord {
					newState.RevealedWord = sState.currentWord
				}
				for i, c := range connections {
					newState.PlayerIndex = i
					if i == s.PlayerIndex {
						newState.Warning = s.Warning
					} else {
						newState.Warning = ""
					}
					if err := c.WriteJSON(newState); err != nil {
						println(err)
					}
				}
			case <-timeoutChannel:
				newState := clientState{
					Players:        sState.players,
					Turn:           sState.turn,
					Host:           sState.curHostIndex,
					RevealedWord:   sState.revealedWord,
					GuessesLeft:    sState.guessesLeft,
					LettersGuessed: sState.guessed,
					NeedNewWord:    sState.needNewWord,
					Warning:        "timed out",
				}
				for i, conn := range connections {
					newState.PlayerIndex = i
					if err := conn.WriteJSON(newState); err != nil {
						println(err)
					}
				}
			}
		}
	}()
	gin.SetMode(gin.ReleaseMode)
	r.Run("localhost:8000") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
