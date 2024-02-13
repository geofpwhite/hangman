package hangman

import (
	"database/sql"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"
)

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

func (gState *gameState) runTicker(timeoutChannel chan int, inputChannel chan inputInfo, closeGameChannel chan int) {
	ticker := time.NewTicker(60 * time.Second)
	timeoutsInARow := 0
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			println("ticker")

			timeoutChannel <- (*gState).gameIndex
			timeoutsInARow++
			if timeoutsInARow >= len(gState.players) {
				closeGameChannel <- gState.gameIndex
			}

			// Send information over the WebSocket connection every 60 seconds
		case x := <-inputChannel:
			fmt.Println("ticker input channel", x)
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
func (gState *gameState) newPlayer(p player) {
	gState.mut.Lock()
	defer gState.mut.Unlock()
	gState.players = append(gState.players, p)
	if len(gState.players) == 2 && !gState.needNewWord {
		gState.turn = 1
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

func (gState *gameState) removePlayer(playerIndex int, tickerInputChannels []chan inputInfo, outputChannel chan clientState, closeGameChannel chan int) {
	// playerIndex := removePlayer[1]
	gState.mut.Lock()
	defer gState.mut.Unlock()
	gState.players = slices.Delete(gState.players, playerIndex, playerIndex+1)

	if len(gState.players) == 0 {
		closeGameChannel <- gState.gameIndex
		// gState.mut.Unlock()
		// gState.closeGame()
		// gState.mut.Lock()
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

}
