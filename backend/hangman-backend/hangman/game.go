package hangman

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"strings"
	"time"
)

const HOST_WINS = 0
const HOST_LOSES = 1

type serverState struct {
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
}

var sState serverState = serverState{}

func runTicker(timeoutChannel chan bool, inputChannel chan info) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case tick := <-ticker.C:
			fmt.Println(tick)
			timeoutChannel <- true
			// Send information over the WebSocket connection every 60 seconds
		case x := <-inputChannel:
			if len(sState.players) == 0 {
				continue
			}
			if x.PlayerIndex == sState.turn {
				ticker.Reset(60 * time.Second)
				fmt.Println("ticker reset")
			}
			// ticker = time.NewTicker(1 * time.Second)
		}
	}
}

func guess(letter rune, outputChannel chan clientState) {
	if sState.needNewWord {
		return
	}
	if !strings.Contains(sState.guessed, string(letter)) {
		good := false
		sState.guessed += string(letter)
		for i, char := range sState.currentWord {
			if char == letter {
				sState.revealedWord = sState.revealedWord[:i] + string(letter) + sState.revealedWord[i+1:]
				good = true
			}
		}
		changedPartsOfState := clientState{}

		if sState.currentWord == sState.revealedWord {
			changedPartsOfState.NeedNewWord = true
			sState.needNewWord = true
			sState.turn = (sState.curHostIndex + 2) % len(sState.players)
			sState.curHostIndex = (sState.curHostIndex + 1) % len(sState.players)
			sState.winner = HOST_LOSES
			changedPartsOfState.Host, changedPartsOfState.Turn = sState.curHostIndex, sState.turn
		} else if sState.guessesLeft == 1 && !good {
			sState.needNewWord = true
			sState.turn = (sState.curHostIndex + 2) % len(sState.players)
			sState.winner = HOST_WINS
			sState.curHostIndex = (sState.curHostIndex + 1) % len(sState.players)
		} else if !good {
			sState.guessesLeft--
			sState.turn = (sState.turn + 1) % len(sState.players)
			if sState.turn == sState.curHostIndex {
				sState.turn = (sState.turn + 1) % len(sState.players)
			}
		}
		outputChannel <- changedPartsOfState
	}
}

func newWord(word string, outputChannel chan clientState) {
	x, _ := sState.wordCheck.Query("select word from words where word='" + word + "'")
	result := ""
	if x.Next() {

		x.Scan(&result)
		fmt.Println(result)
		if result == "" {
			return
		}
	} else {
		return
	}
	sState.currentWord = word
	sState.revealedWord = ""
	sState.needNewWord = false
	sState.guessed = ""
	sState.guessesLeft = 6
	sState.winner = -1
	for range word {
		sState.revealedWord += "_"
	}
	sState.turn = (sState.turn + 1) % len(sState.players)
	if sState.turn == sState.curHostIndex {
		sState.turn = (sState.turn + 1) % len(sState.players)
	}
	outputChannel <- clientState{}
}

func game(inputChannel chan info, timeoutChannel chan bool, outputChannel chan clientState) {
	tickerTimeoutChannel := make(chan (bool))
	tickerInputChannel := make(chan (info))
	go runTicker(tickerTimeoutChannel, tickerInputChannel)
	for {
		select {
		case timeout := <-tickerTimeoutChannel:
			println("timeout")
			//timed out, move to the next player
			if len(sState.players) == 0 {
				continue
			}
			sState.turn = (sState.turn + 1) % len(sState.players)
			if sState.curHostIndex == sState.turn {
				sState.turn = (sState.turn + 1) % len(sState.players)
			}
			timeoutChannel <- timeout //for the websocket to update everybody

		case info := <-inputChannel:
			tickerInputChannel <- info //for the ticker to handle
			//main part of loop
			// determine from the info object what kind of action that user is gonna take
			if len(info.Guess) >= 1 {
				//they sState.guessed a letter that's one character
				if info.PlayerIndex == sState.turn {
					guess(rune(info.Guess[0]), outputChannel)
				} else {
					outputChannel <- clientState{Warning: "not your turn", PlayerIndex: info.PlayerIndex}
				}
			} else if len(info.Word) > 0 { //they are giving us a new word
				if sState.needNewWord && info.PlayerIndex == sState.curHostIndex {
					newWord(info.Word, outputChannel)
				} else {
					outputChannel <- clientState{Warning: "you can't pick the word right now", PlayerIndex: info.PlayerIndex}
				}
			} else if len(info.Username) > 0 {
				sState.players[info.PlayerIndex] = info.Username
				outputChannel <- clientState{}
			}
		}

	}
}
