package hangman

import (
	"database/sql"
	"fmt"
	"log"
	"slices"
	"strings"
	"sync"
	"time"
)

type chatLog struct {
	Message string `json:"message"`
	Sender  string `json:"sender"`
}

/*
struct containing necessary fields for game to run
*/
type gameState struct {
	wordCheck           *sql.DB
	currentWord         string
	revealedWord        string
	guessed             string
	players             []player
	curHostIndex        int
	turn                int
	guessesLeft         int
	needNewWord         bool
	winner              int
	mut                 *sync.Mutex
	chatLogs            []chatLog
	consecutiveTimeouts int
	randomlyChosen      bool //boolean for methods to check if they need to act differently because the backend randomly chose a word
	gameHash            string
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
		mut:          &sync.Mutex{},
	}
	gameCode := Hash(6)
	gameHashes[gameCode] = gState
	gState.gameHash = gameCode
	return gState
}

/*
starts a ticker that either times out the current turn and increments it, or resets back to 0 on user input
*/
func (gState *gameState) runTicker(timeoutChannel chan string, inputChannel chan inputInfo, closeGameChannel chan string) {
	ticker := time.NewTicker(60 * time.Second)
	gState.consecutiveTimeouts = 0
	defer ticker.Stop()
	defer close(inputChannel) // this may be bad practice to close from the reader side but

	for {
		select {
		case <-ticker.C:
			log.Println("ticker")

			timeoutChannel <- (*gState).gameHash
			gState.consecutiveTimeouts++
			if gState.consecutiveTimeouts >= len(gState.players) {
				closeGameChannel <- gState.gameHash
			}

		case x := <-inputChannel:
			log.Println("ticker input channel", x)
			if len((*gState).players) == 0 || x.PlayerIndex == -1 {
				return
			}
			if x.PlayerIndex == (*gState).turn {
				ticker.Stop()
				ticker = time.NewTicker(60 * time.Second)
				fmt.Println("ticker reset")
				gState.consecutiveTimeouts = 0
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

func (gState *gameState) guess(letter rune) (bool, clientState) {
	gState.mut.Lock()
	defer gState.mut.Unlock()
	if gState.needNewWord {
		return false, clientState{}
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
		changedPartsOfState := clientState{GameHash: gState.gameHash}

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
			if gState.turn == gState.curHostIndex && !gState.randomlyChosen {
				gState.turn = (gState.turn + 1) % len(gState.players)
			}
		}
		return true, changedPartsOfState
	}
	return false, clientState{}
}

func (gState *gameState) randomNewWord() {
	gState.mut.Lock()
	defer gState.mut.Unlock()
	x, _ := gState.wordCheck.Query("SELECT word FROM words WHERE LENGTH(word)>5 and word not like '%-%' ORDER BY RANDOM() LIMIT 1;")
	result := ""
	if x.Next() {
		x.Scan(&result)
		if result == "" {
			return
		}
	} else {
		return
	}
	gState.currentWord = result
	gState.revealedWord = ""
	gState.needNewWord = false
	gState.guessed = ""
	gState.guessesLeft = 6
	gState.winner = -1
	gState.randomlyChosen = true
	for range result {
		gState.revealedWord += "_"
	}
	gState.turn = (gState.curHostIndex + 1) % len(gState.players)
}

func (gState *gameState) newWord(word string) {
	gState.mut.Lock()
	defer gState.mut.Unlock()
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
	gState.currentWord = word
	gState.revealedWord = ""
	gState.needNewWord = false
	gState.guessed = ""
	gState.guessesLeft = 6
	gState.winner = -1
	gState.randomlyChosen = false
	for range word {
		gState.revealedWord += "_"
	}
	gState.turn = (gState.curHostIndex + 1) % len(gState.players)
}

func (gState *gameState) closeGame() {
	gState.mut.Lock()
	defer gState.mut.Unlock()
	for _, p := range gState.players {
		delete(hashes, p.hash)
	}
	delete(gameHashes, gState.gameHash)
}

func (gState *gameState) removePlayer(playerIndex int) {
	if len(gState.players) == 0 {
		return
	}
	gState.mut.Lock()
	defer gState.mut.Unlock()
	gState.players = slices.Delete(gState.players, playerIndex, playerIndex+1)
	if len(gState.players) == 0 {
		return
	}
	gState.turn = gState.turn % len(gState.players)
	gState.curHostIndex = gState.curHostIndex % len(gState.players)
	if gState.needNewWord && gState.curHostIndex != gState.turn {
		gState.turn = gState.curHostIndex
	} else if !gState.needNewWord && gState.curHostIndex == gState.turn {
		gState.turn = (gState.turn + 1) % len(gState.players)
	}
}

func (gState *gameState) handleTickerTimeout() string {
	gState.mut.Lock()
	defer gState.mut.Unlock()
	if gState.needNewWord {
		gState.curHostIndex = (gState.curHostIndex + 1) % len(gState.players)
		gState.turn = gState.curHostIndex
	} else {
		(*gState).turn = ((*gState).turn + 1) % len((*gState).players)
		if (*gState).curHostIndex == (*gState).turn {
			(*gState).turn = ((*gState).turn + 1) % len((*gState).players)
		}
	}
	return gState.gameHash
}

func (gState *gameState) changeUsername(playerIndex int, newUsername string) {
	log.Println("change username")
	gState.mut.Lock()
	defer gState.mut.Unlock()
	if slices.IndexFunc(gState.players, func(p player) bool {
		return p.username == newUsername
	}) == -1 {
		oldUsername := gState.players[playerIndex].username
		gState.players[playerIndex].username = newUsername
		for i, chat := range gState.chatLogs {
			if chat.Sender == oldUsername {
				gState.chatLogs[i].Sender = newUsername
			}
		}
	}
}

func (gState *gameState) chat(message string, playerIndex int) {
	gState.mut.Lock()
	defer gState.mut.Unlock()
	gState.chatLogs = append(gState.chatLogs,
		chatLog{
			Message: message,
			Sender:  gState.players[playerIndex].username,
		})
}
