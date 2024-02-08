package hangman

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
)

const HOST_WINS = 0
const HOST_LOSES = 1

type gameState struct {
	currentWord  string
	revealedWord string
	guessed      string
	players      []string
	curHostIndex int
	turn         int
	guessesLeft  int
	needNewWord  bool
	winner       int
	wordCheck    *sql.DB
	gameIndex    int
	connections  []*websocket.Conn
}

// var gState serverState = serverState{}
var gStates []*gameState = []*gameState{}

func (gState *gameState) runTicker(timeoutChannel chan int, inputChannel chan inputInfo) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

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
	x, _ := (*gState).wordCheck.Query("select word from words where word='" + word + "'")
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
	(*gState).turn = ((*gState).turn + 1) % len((*gState).players)
	if (*gState).turn == (*gState).curHostIndex {
		(*gState).turn = ((*gState).turn + 1) % len((*gState).players)
	}
	outputChannel <- clientState{GameIndex: (*gState).gameIndex}
}

func newGame() *gameState {
	gState := new(gameState)

	(*gState).currentWord = ""
	(*gState).revealedWord = ""

	(*gState).winner = -1
	wordCheck, _ := sql.Open("sqlite3", "./words.db")
	(*gState).wordCheck = wordCheck
	(*gState).needNewWord = true
	(*gState).guessesLeft = 6
	(*gState).players = make([]string, 0)
	(*gState).gameIndex = len(gStates)
	gStates = append(gStates, gState)
	return gState
}

func (gState *gameState) closeGame() {
	for i := range gStates[(*gState).gameIndex+1:] {
		gStates[i+(*gState).gameIndex+1].gameIndex--
	}
	gStates = append(gStates[:(*gState).gameIndex], gStates[(*gState).gameIndex+1:]...)

}

func game(inputChannel chan inputInfo, timeoutChannel chan int, outputChannel chan clientState, newGameChannel chan bool, closeGameChannel chan int) {
	tickerInputChannels := []chan inputInfo{}
	tickerTimeoutChannel := make(chan (int))
	tickerInputChannel := make(chan (inputInfo))
	tickerInputChannels = append(tickerInputChannels, tickerInputChannel)
	go gStates[0].runTicker(tickerTimeoutChannel, tickerInputChannel)
	for {
		select {

		case gameIndex := <-closeGameChannel:
			gStates[gameIndex].closeGame()
			tickerInputChannels[gameIndex] <- inputInfo{GameIndex: -1}
			tickerInputChannels = append(tickerInputChannels[:gameIndex], tickerInputChannels[gameIndex+1:]...)

		case <-newGameChannel:
			gState := newGame()
			newTickerInputChannel := make(chan (inputInfo))
			tickerInputChannels = append(tickerInputChannels, newTickerInputChannel)
			go (*gState).runTicker(tickerTimeoutChannel, newTickerInputChannel)

		case gameIndex := <-tickerTimeoutChannel:
			if gameIndex >= len(gStates) {
				closeGameChannel <- gameIndex
			}
			gState := &gStates[gameIndex]
			println("timeout")
			//timed out, move to the next player
			if len((*gState).players) == 0 {
				continue
			}
			(*gState).turn = ((*gState).turn + 1) % len((*gState).players)
			if (*gState).curHostIndex == (*gState).turn {
				(*gState).turn = ((*gState).turn + 1) % len((*gState).players)
			}
			timeoutChannel <- gameIndex //for the websocket to update everybody

		case info := <-inputChannel:
			gState := &gStates[info.GameIndex]
			fmt.Println(info)
			fmt.Println(gState)
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
				(*gState).players[info.PlayerIndex] = info.Username
				outputChannel <- clientState{}
			}
		}

	}
}
