package hangman

import (
	"database/sql"
	"fmt"
	// "fmt"
	"slices"
	"strings"
	"sync"
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
	wordCheck    *sql.DB
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
	mut          *sync.Mutex
}

// var gState serverState = serverState{}
var gStates []*gameState = []*gameState{}

func (gState *gameState) runTicker(timeoutChannel chan int, inputChannel chan inputInfo, closeGameChannel chan int) {
	ticker := time.NewTicker(60 * time.Second)
	timeoutsInARow := 0
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:

			timeoutChannel <- (*gState).gameIndex
			timeoutsInARow++
			if timeoutsInARow >= len(gState.players) {
				closeGameChannel <- gState.gameIndex
			}

			// Send information over the WebSocket connection every 60 seconds
		case x := <-inputChannel:
			if len((*gState).players) == 0 || x.PlayerIndex == -1 {
				return
			}
			if x.PlayerIndex == (*gState).turn {
				ticker.Reset(60 * time.Second)
				timeoutsInARow = 0
			}
			// ticker = time.NewTicker(1 * time.Second)
		}
	}
}

func (gState *gameState) guess(letter rune, outputChannel chan clientState) {
	gState.mut.Lock()
	defer gState.mut.Unlock()
	if gState.needNewWord {
		return
	}
	if !strings.Contains(gState.guessed, string(letter)) {
		good := false
		gState.guessed += string(letter)
		for i, char := range gState.currentWord {
			if char == letter {
				gState.revealedWord = gState.revealedWord[:i] + string(letter) + gState.revealedWord[i+1:]
				good = true
			}
		}
		changedPartsOfState := clientState{GameIndex: gState.gameIndex}

		if gState.currentWord == gState.revealedWord {
			changedPartsOfState.NeedNewWord = true
			gState.needNewWord = true
			gState.turn = (gState.curHostIndex + 2) % len(gState.players)
			gState.curHostIndex = (gState.curHostIndex + 1) % len(gState.players)
			gState.winner = HOST_LOSES
			changedPartsOfState.Host, changedPartsOfState.Turn = gState.curHostIndex, gState.turn
		} else if gState.guessesLeft == 1 && !good {
			gState.needNewWord = true
			gState.turn = (gState.curHostIndex + 2) % len(gState.players)
			gState.winner = HOST_WINS
			gState.curHostIndex = (gState.curHostIndex + 1) % len(gState.players)
		} else if !good {
			gState.guessesLeft--
			gState.turn = (gState.turn + 1) % len(gState.players)
			if gState.turn == gState.curHostIndex {
				gState.turn = (gState.turn + 1) % len(gState.players)
			}
		}
		outputChannel <- changedPartsOfState
	}
}

func (gState *gameState) newWord(word string, outputChannel chan clientState) {
	fmt.Println("new word")
	x, _ := gState.wordCheck.Query("select word from words where word='" + word + "'")
	result := ""
	if x.Next() {

		x.Scan(&result)
		if result == "" {
			return
		}
	} else {
		return
	}
	gState.mut.Lock()
	defer gState.mut.Unlock()
	gState.currentWord = word
	gState.revealedWord = ""
	gState.needNewWord = false
	gState.guessed = ""
	gState.guessesLeft = 6
	gState.winner = -1
	for range word {
		gState.revealedWord += "_"
	}
	gState.turn = (gState.curHostIndex + 1) % len(gState.players)
	outputChannel <- clientState{GameIndex: gState.gameIndex}
}

func newGame() *gameState {
	wordCheck, _ := sql.Open("sqlite3", "./words.db")
	gState := &gameState{
		wordCheck:    wordCheck,
		currentWord:  "",
		revealedWord: "",
		winner:       -1,
		needNewWord:  true,
		guessesLeft:  6,
		players:      make([]player, 0),
		gameIndex:    len(gStates),
		mut:          &sync.Mutex{},
	}

	gStates = append(gStates, gState)
	return gState
}

func (gState *gameState) closeGame() {
	gState.mut.Lock()
	defer gState.mut.Unlock()
	for i := range gStates[gState.gameIndex+1:] {
		gStates[i+gState.gameIndex+1].gameIndex--
	}

	gStates = slices.Delete(gStates, gState.gameIndex, gState.gameIndex+1)

}

func game(
	inputChannel chan input,
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
	go gStates[0].runTicker(tickerTimeoutChannel, tickerInputChannels[0], closeGameChannel)
	go func() {
		for gameIndex := range closeGameChannel {
			if gameIndex < len(gStates) {
				gStates[gameIndex].gameIndex = gameIndex
				gStates[gameIndex].closeGame()
				println("game closed")
				tickerInputChannels[gameIndex] <- inputInfo{GameIndex: -1}
				close(tickerInputChannels[gameIndex])
				tickerInputChannels = slices.Delete(tickerInputChannels, gameIndex, gameIndex+1)
			}
		}
	}()
	for {
		select {
		case removePlayer := <-removePlayerChannel:
			if removePlayer[0] >= len(gStates) {
				continue
			}
			gState := gStates[removePlayer[0]]
			if removePlayer[1] >= len(gState.players) {
				continue
			}
			playerIndex := removePlayer[1]
			gState.mut.Lock()
			gState.players = slices.Delete(gState.players, playerIndex, playerIndex+1)
			// println(len(gState.players))

			if len(gState.players) == 0 {
				// closeGameChannel <- gState.gameIndex
				gState.mut.Unlock()
				gState.closeGame()
				gState.mut.Lock()
				tickerInputChannels[gState.gameIndex] <- inputInfo{GameIndex: -1}
				close(tickerInputChannels[gState.gameIndex])
				tickerInputChannels = slices.Delete(tickerInputChannels, gState.gameIndex, gState.gameIndex+1)
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
			gState.mut.Unlock()

		/* case gameIndex := <-closeGameChannel:
		gStates[gameIndex].gameIndex = gameIndex
		gStates[gameIndex].closeGame()
		println("game closed")
		tickerInputChannels[gameIndex] <- inputInfo{GameIndex: -1}
		close(tickerInputChannels[gameIndex])
		tickerInputChannels = slices.Delete(tickerInputChannels, gameIndex, gameIndex+1)
		*/
		case <-newGameChannel:
			gState := newGame()
			println("newgame")
			newTickerInputChannel := make(chan (inputInfo))
			tickerInputChannels = append(tickerInputChannels, newTickerInputChannel)
			go (*gState).runTicker(tickerTimeoutChannel, newTickerInputChannel, closeGameChannel)

		case gameIndex := <-tickerTimeoutChannel:
			if gameIndex >= len(gStates) {
				continue
			}
			gState := gStates[gameIndex]
			println("timeout")
			//timed out, move to the next player
			if len((*gState).players) <= 1 {
				closeGameChannel <- gameIndex
				continue
			}
			gState.mut.Lock()
			if gState.needNewWord {
				gState.curHostIndex = (gState.curHostIndex + 1) % len(gState.players)
				gState.turn = gState.curHostIndex
			} else {

				(*gState).turn = ((*gState).turn + 1) % len((*gState).players)
				if (*gState).curHostIndex == (*gState).turn {
					(*gState).turn = ((*gState).turn + 1) % len((*gState).players)
				}
			}
			gState.mut.Unlock()
			timeoutChannel <- gameIndex //for the websocket to update everybody

		case info := <-inputChannel:
			info.ChangeStateAccordingToInput(outputChannel)
			fmt.Println("changed state")
		}

	}
}
