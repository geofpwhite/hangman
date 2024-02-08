package hangman

import (
	"fmt"
	"log"
	"net/http"
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

func (gState *gameState) handleWebSocket(conn *websocket.Conn, inputChannel chan inputInfo, timeoutChannel chan int, outputChannel chan clientState, closeGameChannel chan int) {
	// first thing to do when we have a new connection is tell them the current game state

	gState.connections = append(gState.connections, conn)
	playerIndex := len(gState.players)
	gState.players = append(gState.players, "Player "+strconv.Itoa(playerIndex+1))
	if playerIndex == 1 {
		gState.turn = 1
	}
	defer func() {
		if len(gState.players)-1 == playerIndex {
			gState.players, gState.connections = gState.players[:playerIndex], gState.connections[:playerIndex]
		} else {
			gState.players = append(gState.players[:playerIndex], gState.players[playerIndex+1:]...)
			gState.connections = append(gState.connections[:playerIndex], gState.connections[playerIndex+1:]...)
		}
		if len(gState.connections) == 0 {
			closeGameChannel <- gState.gameIndex
		} else {
			outputChannel <- clientState{GameIndex: gState.gameIndex}
		}
	}()

	currentState := clientState{
		Players:        gState.players,
		Turn:           gState.turn,
		Host:           gState.curHostIndex,
		RevealedWord:   gState.revealedWord,
		GuessesLeft:    gState.guessesLeft,
		LettersGuessed: gState.guessed,
		NeedNewWord:    gState.needNewWord,
		Warning:        "",
		PlayerIndex:    playerIndex,
	}
	fmt.Println(currentState)
	for i, c := range gState.connections {
		currentState.PlayerIndex = i
		c.WriteJSON(currentState)
	}

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			print("er")
			continue
		}
		i := inputInfo{
			GameIndex:   gState.gameIndex,
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

func server(inputChannel chan inputInfo, timeoutChannel chan int, outputChannel chan clientState, newGameChannel chan bool, closeGameChannel chan int) {
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
	r.GET("/new_game", func(c *gin.Context) {
		newGameChannel <- true

		c.String(200, "ok")
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
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()

		gStates[gameIndex].handleWebSocket(conn, inputChannel, timeoutChannel, outputChannel, closeGameChannel)
		// Handle WebSocket connections here
	})
	go func() {
		for {
			// write changed state to clients
			select {
			case s := <-outputChannel:

				gState := &gStates[s.GameIndex]

				fmt.Println(s.GameIndex, " game index")
				newState := clientState{
					GameIndex:      (*gState).gameIndex,
					Players:        (*gState).players,
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
				fmt.Println(gState, " gState")
				for i, c := range (*gState).connections {
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
			case gameIndex := <-timeoutChannel:
				gState := &gStates[gameIndex]
				fmt.Println(gameIndex, " game index")
				newState := clientState{
					Players:        (*gState).players,
					Turn:           (*gState).turn,
					Host:           (*gState).curHostIndex,
					RevealedWord:   (*gState).revealedWord,
					GuessesLeft:    (*gState).guessesLeft,
					LettersGuessed: (*gState).guessed,
					NeedNewWord:    (*gState).needNewWord,
					GameIndex:      (*gState).gameIndex,
					Warning:        "timed out",
				}

				for i, conn := range (*gState).connections {
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
