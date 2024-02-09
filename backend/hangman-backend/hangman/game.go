package hangman

import (
	"database/sql"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
)

const HOST_WINS = 1
const HOST_LOSES = 2

type player struct {
	username   string
	connection *websocket.Conn
}

type gameState struct {
	currentWord  string
	revealedWord string
	guessed      string
	players      []player
	curHostIndex int
	turn         int
	guessesLeft  int
	needNewWord  bool
	winner       int
	gameIndex    int
	// connections  []*websocket.Conn
}

var wordCheck *sql.DB

// var gState serverState = serverState{}
var gStates []*gameState = []*gameState{}

func (gState *gameState) runTicker(timeoutChannel chan int, inputChannel chan inputInfo) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	defer close(timeoutChannel)

	for {
		select {
		case tick := <-ticker.C:
			fmt.Println(tick)
			timeoutChannel <- (*gState).gameIndex
			// Send information over the WebSocket connection every 60 seconds
		case x := <-inputChannel:
			if len((*gState).players) == 0 {
				continue
			}
			if x.PlayerIndex == (*gState).turn {
				ticker.Reset(60 * time.Second)
				fmt.Println("ticker reset")
			}
			// ticker = time.NewTicker(1 * time.Second)
		}
	}
}

func (gState *gameState) guess(letter rune, outputChannel chan clientState) {
	if (*gState).needNewWord {
		return
	}
	if !strings.Contains((*gState).guessed, string(letter)) {
		good := false
		(*gState).guessed += string(letter)
		for i, char := range (*gState).currentWord {
			if char == letter {
				(*gState).revealedWord = (*gState).revealedWord[:i] + string(letter) + (*gState).revealedWord[i+1:]
				good = true
			}
		}
		changedPartsOfState := clientState{GameIndex: (*gState).gameIndex}

		if (*gState).currentWord == (*gState).revealedWord {
			changedPartsOfState.NeedNewWord = true
			(*gState).needNewWord = true
			(*gState).turn = ((*gState).curHostIndex + 2) % len((*gState).players)
			(*gState).curHostIndex = ((*gState).curHostIndex + 1) % len((*gState).players)
			(*gState).winner = HOST_LOSES
			changedPartsOfState.Host, changedPartsOfState.Turn = (*gState).curHostIndex, (*gState).turn
		} else if (*gState).guessesLeft == 1 && !good {
			(*gState).needNewWord = true
			(*gState).turn = ((*gState).curHostIndex + 2) % len((*gState).players)
			(*gState).winner = HOST_WINS
			(*gState).curHostIndex = ((*gState).curHostIndex + 1) % len((*gState).players)
		} else if !good {
			(*gState).guessesLeft--
			(*gState).turn = ((*gState).turn + 1) % len((*gState).players)
			if (*gState).turn == (*gState).curHostIndex {
				(*gState).turn = ((*gState).turn + 1) % len((*gState).players)
			}
		}
		outputChannel <- changedPartsOfState
	}
}

func (gState *gameState) newWord(word string, outputChannel chan clientState) {
	x, _ := wordCheck.Query("select word from words where word='" + word + "'")
	fmt.Println(x)
	result := ""
	if x.Next() {

		x.Scan(&result)
		fmt.Println(result, " result")
		if result == "" {
			return
		}
	} else {
		return
	}
	(*gState).currentWord = word
	(*gState).revealedWord = ""
	(*gState).needNewWord = false
	(*gState).guessed = ""
	(*gState).guessesLeft = 6
	(*gState).winner = -1
	for range word {
		(*gState).revealedWord += "_"
	}
	(*gState).turn = ((*gState).curHostIndex + 1) % len((*gState).players)
	outputChannel <- clientState{GameIndex: (*gState).gameIndex}
}

func newGame() *gameState {
	gState := new(gameState)

	gState.currentWord = ""
	gState.revealedWord = ""

	gState.winner = -1
	gState.needNewWord = true
	gState.guessesLeft = 6
	gState.players = make([]player, 0)
	gState.gameIndex = len(gStates)
	gStates = append(gStates, gState)
	return gState
}

func (gState *gameState) closeGame() {
	fmt.Println("closing game")
	for _, gs := range gStates {

		fmt.Println(gs)
	}
	for i := range gStates[gState.gameIndex+1:] {
		gStates[i+gState.gameIndex+1].gameIndex--
	}
	gStates = append(gStates[:gState.gameIndex], gStates[gState.gameIndex+1:]...)

}

func game(
	inputChannel chan inputInfo,
	timeoutChannel chan int,
	outputChannel chan clientState,
	newGameChannel chan bool,
	closeGameChannel chan int,
	removePlayerChannel chan [2]int,
) {
	tickerInputChannels := []chan inputInfo{}
	tickerTimeoutChannel := make(chan (int))
	{
		tickerInputChannel := make(chan (inputInfo))
		tickerInputChannels = append(tickerInputChannels, tickerInputChannel)
	}
	go gStates[0].runTicker(tickerTimeoutChannel, tickerInputChannels[0])
	for {
		select {
		case removePlayer := <-removePlayerChannel:
			gState := gStates[removePlayer[0]]
			playerIndex := removePlayer[1]
			fmt.Println("Removing ", gState.players[playerIndex], ", they are index number ", playerIndex)
			println(len(gState.players))
			gState.players = slices.Delete(gState.players, playerIndex, playerIndex+1)
			println(len(gState.players))

			if len(gState.players) == 0 {
				// closeGameChannel <- gState.gameIndex
				gState.closeGame()
			} else {
				// outputChannel <- clientState{GameIndex: gState.gameIndex}
				gState.turn = gState.turn % len(gState.players)
				gState.curHostIndex = gState.curHostIndex % len(gState.players)
				if gState.needNewWord && gState.curHostIndex != gState.turn {
					gState.turn = gState.curHostIndex
				} else if !gState.needNewWord && gState.curHostIndex == gState.turn {
					gState.turn = (gState.turn + 1) % len(gState.players)
				}
				outputChannel <- clientState{GameIndex: gState.gameIndex}
			}

		case gameIndex := <-closeGameChannel:
			gStates[gameIndex].gameIndex = gameIndex
			gStates[gameIndex].closeGame()
			println("game closed")
			tickerInputChannels[gameIndex] <- inputInfo{GameIndex: -1}
			tickerInputChannels = append(tickerInputChannels[:gameIndex], tickerInputChannels[gameIndex+1:]...)

		case <-newGameChannel:
			gState := newGame()
			println("newgame")
			newTickerInputChannel := make(chan (inputInfo))
			tickerInputChannels = append(tickerInputChannels, newTickerInputChannel)
			go (*gState).runTicker(tickerTimeoutChannel, newTickerInputChannel)

		case gameIndex := <-tickerTimeoutChannel:
			if gameIndex >= len(gStates) {
				closeGameChannel <- gameIndex
			}
			gState := gStates[gameIndex]
			println("timeout")
			//timed out, move to the next player
			if len((*gState).players) <= 1 {
				closeGameChannel <- gameIndex
			}
			if gState.needNewWord {
				gState.curHostIndex = (gState.curHostIndex + 1) % len(gState.players)
				gState.turn = gState.curHostIndex
			} else {

				(*gState).turn = ((*gState).turn + 1) % len((*gState).players)
				if (*gState).curHostIndex == (*gState).turn {
					(*gState).turn = ((*gState).turn + 1) % len((*gState).players)
				}
			}
			timeoutChannel <- gameIndex //for the websocket to update everybody

		case info := <-inputChannel:
			gState := &gStates[info.GameIndex]
			fmt.Println(info)
			for _, gs := range gStates {

				fmt.Println(gs)
			}
			tickerInputChannels[info.GameIndex] <- info //for the ticker to handle
			//main part of loop
			// determine from the info object what kind of action that user is gonna take
			if len(info.Guess) >= 1 {
				//they (*gState).guessed a letter that's one character
				if info.PlayerIndex == (*gState).turn {
					fmt.Println("guess")
					(*gState).guess(rune(info.Guess[0]), outputChannel)
				} else {
					outputChannel <- clientState{Warning: "not your turn", PlayerIndex: info.PlayerIndex}
				}
			} else if len(info.Word) > 0 { //they are giving us a new word

				fmt.Println("word")
				fmt.Println(info.Word)
				fmt.Println((*gState).needNewWord, info.PlayerIndex, (*gState).curHostIndex)

				if (*gState).needNewWord && info.PlayerIndex == (*gState).curHostIndex {
					(*gState).newWord(info.Word, outputChannel)
				} else {
					outputChannel <- clientState{Warning: "you can't pick the word right now", PlayerIndex: info.PlayerIndex}
				}
			} else if len(info.Username) > 0 {
				(*gState).players[info.PlayerIndex].username = info.Username
				outputChannel <- clientState{}
			}

		}

	}
}
