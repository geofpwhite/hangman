package hangman

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type info struct {
	Username    string `json:"username"`
	Guess       string `json:"guess"`
	Word        string `json:"word"`
	Signup      bool   `json:"signup"`
	PlayerIndex int    `json:"playerIndex"`
}

var db *sql.DB

type clientState struct {
	Players        []string `json:"players"`
	Turn           int      `json:"turn"`
	Host           int      `json:"host"`
	RevealedWord   string   `json:"revealedWord"`
	GuessesLeft    int      `json:"guessesLeft"`
	LettersGuessed string   `json:"lettersGuessed"`
	NeedNewWord    bool     `json:"needNewWord"`
	Warning        string   `json:"warning"`
	PlayerIndex    int      `json:"playerIndex"` // changes for each connection that the update state object is sent to
	Winner         int      `json:"winner"`
}

func Run() {
	sState.currentWord = ""
	sState.revealedWord = ""

	defer db.Close()
	sState.winner = -1
	wordCheck, _ := sql.Open("sqlite3", "./words.db")
	sState.wordCheck = wordCheck

	sState.players = make([]string, 0)
	inputChannel := make(chan (info))
	outputChannel := make(chan (clientState))
	timeoutChannel := make(chan (bool))
	sState.needNewWord = true
	sState.guessesLeft = 6
	/* defer close(inputChannel)
	defer close(outputChannel)
	defer close(timeoutChannel) */
	go game(inputChannel, timeoutChannel, outputChannel)
	// go newView(cliChannel)
	server(inputChannel, timeoutChannel, outputChannel)
}
