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

func handleWebSocket(conn *websocket.Conn, inputChannel chan info, timeoutChannel chan bool, outputChannel chan state) {
	//first thing to do when we have a new connection is tell them the current game state

	connections = append(connections, conn)
	playerIndex := len(players)
	players = append(players, "Player "+strconv.Itoa(playerIndex+1))
	defer func() {
		if len(players)-1 == playerIndex {
			players, connections = players[:playerIndex], connections[:playerIndex]

		} else {
			players = append(players[:playerIndex], players[playerIndex+1:]...)
			connections = append(connections[:playerIndex], connections[playerIndex+1:]...)
		}
		outputChannel <- state{}
	}()

	currentState := state{
		Players:        players,
		Turn:           turn,
		Host:           curHostIndex,
		RevealedWord:   revealedWord,
		GuessesLeft:    guessesLeft,
		LettersGuessed: guessed,
		NeedNewWord:    needNewWord,
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
			switch {
			case string(p)[:2] == "g:":
				i.Guess = string(p[2:])

			case string(p)[:2] == "u:":
				i.Username = string(p)[2:]
			case string(p)[:2] == "w:":
				i.Word = string(p)[2:]
			default:
				continue
			}

			inputChannel <- i
			//send what the user wants to do to input channel for game function to handle

		}

	}
}

func server(inputChannel chan info, timeoutChannel chan bool, outputChannel chan state) {

	r := gin.Default()
	gin.SetMode(gin.ReleaseMode)
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
			//write changed state to clients
			select {
			case s := <-outputChannel:
				println("output")
				println(turn)
				newState := state{
					Players:        players,
					Turn:           turn,
					Host:           curHostIndex,
					RevealedWord:   revealedWord,
					GuessesLeft:    guessesLeft,
					LettersGuessed: guessed,
					NeedNewWord:    needNewWord,
					Warning:        "",
					Winner:         winner,
				}
				if newState.NeedNewWord {
					newState.RevealedWord = currentWord
				}
				fmt.Println(newState)
				for i, c := range connections {
					newState.PlayerIndex = i
					if i == s.PlayerIndex {
						newState.Warning = s.Warning
					}
					if err := c.WriteJSON(newState); err != nil {
						println(err)
					}
				}
			case <-timeoutChannel:
				newState := state{
					Players:        players,
					Turn:           turn,
					Host:           curHostIndex,
					RevealedWord:   revealedWord,
					GuessesLeft:    guessesLeft,
					LettersGuessed: guessed,
					NeedNewWord:    needNewWord,
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
	r.Run("localhost:8000") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
