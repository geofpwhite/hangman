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

type state struct {
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
	currentWord = ""
	revealedWord = ""
	guessed := make(map[rune]bool)

	defer db.Close()

	for _, _rune := range "abcdefghijklmnopqrstuvwxyz" {
		guessed[_rune] = false
	}
	players = make([]string, 0)
	inputChannel := make(chan (info))
	outputChannel := make(chan (state))
	timeoutChannel := make(chan (bool))
	needNewWord = true
	guessesLeft = 10
	/* defer close(inputChannel)
	defer close(outputChannel)
	defer close(timeoutChannel) */
	go game(inputChannel, timeoutChannel, outputChannel)
	server(inputChannel, timeoutChannel, outputChannel)
}
