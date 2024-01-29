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

var (
	currentWord  string
	revealedWord string
	guessed      string
	players      []string
	curHostIndex int
	turn         int
	guessesLeft  int
	needNewWord  bool = false
	winner       int  = -1
	wordCheck, _      = sql.Open("sqlite3", "./words.db")
)

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
			if len(players) == 0 {
				continue
			}
			if x.PlayerIndex == turn {
				ticker.Reset(60 * time.Second)
				fmt.Println("ticker reset")
			}
			// ticker = time.NewTicker(1 * time.Second)
		}
	}
}

func guess(letter rune, outputChannel chan state) {
	if needNewWord {
		return
	}
	if !strings.Contains(guessed, string(letter)) {
		good := false
		guessed += string(letter)
		for i, char := range currentWord {
			if char == letter {
				revealedWord = revealedWord[:i] + string(letter) + revealedWord[i+1:]
				good = true
			}
		}
		changedPartsOfState := state{}

		if currentWord == revealedWord {
			changedPartsOfState.NeedNewWord = true
			needNewWord = true
			turn = (curHostIndex + 2) % len(players)
			curHostIndex = (curHostIndex + 1) % len(players)
			winner = HOST_LOSES
			changedPartsOfState.Host, changedPartsOfState.Turn = curHostIndex, turn
		} else if guessesLeft == 1 && !good {
			needNewWord = true
			turn = (curHostIndex + 2) % len(players)
			winner = HOST_WINS
			curHostIndex = (curHostIndex + 1) % len(players)
		} else if !good {
			guessesLeft--
			turn = (turn + 1) % len(players)
			if turn == curHostIndex {
				turn = (turn + 1) % len(players)
			}
		}
		outputChannel <- changedPartsOfState
	}
}

func newWord(word string, outputChannel chan state) {
	x, _ := wordCheck.Query("select word from words where word='" + word + "'")
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
	currentWord = word
	revealedWord = ""
	needNewWord = false
	guessed = ""
	guessesLeft = 10
	winner = -1
	for range word {
		revealedWord += "_"
	}
	turn = (turn + 1) % len(players)
	if turn == curHostIndex {
		turn = (turn + 1) % len(players)
	}
	outputChannel <- state{}
}

func game(inputChannel chan info, timeoutChannel chan bool, outputChannel chan state) {
	tickerTimeoutChannel := make(chan (bool))
	tickerInputChannel := make(chan (info))
	go runTicker(tickerTimeoutChannel, tickerInputChannel)
	for {
		select {
		case timeout := <-tickerTimeoutChannel:
			println("timeout")
			//timed out, move to the next player
			if len(players) == 0 {
				continue
			}
			turn = (turn + 1) % len(players)
			if curHostIndex == turn {
				turn = (turn + 1) % len(players)
			}
			timeoutChannel <- timeout //for the websocket to update everybody

		case info := <-inputChannel:
			tickerInputChannel <- info //for the ticker to handle
			//main part of loop
			// determine from the info object what kind of action that user is gonna take
			if len(info.Guess) >= 1 {
				//they guessed a letter that's one character
				if info.PlayerIndex == turn {
					guess(rune(info.Guess[0]), outputChannel)
				} else {
					outputChannel <- state{Warning: "not your turn", PlayerIndex: info.PlayerIndex}
				}
			} else if len(info.Word) > 0 { //they are giving us a new word
				if needNewWord && info.PlayerIndex == curHostIndex {
					newWord(info.Word, outputChannel)
				} else {
					outputChannel <- state{Warning: "you can't pick the word right now", PlayerIndex: info.PlayerIndex}
				}
			} else if len(info.Username) > 0 {
				players[info.PlayerIndex] = info.Username
				outputChannel <- state{Players: players}
			}
		}

	}
}
