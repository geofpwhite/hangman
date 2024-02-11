package hangman

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type inputInfo struct {
	Username    string `json:"username"`
	Guess       string `json:"guess"`
	Word        string `json:"word"`
	Signup      bool   `json:"signup"`
	PlayerIndex int    `json:"playerIndex"`
	GameIndex   int    `json:"gameIndex"`
}

/*  */
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
	GameIndex      int      `json:"gameIndex"`
}

func Run() {
	{ //outside of scope so it can be garbage collected.... I think
		gState := newGame()
		gState.players = make([]player, 0)
	}
	inputChannel := make(chan (inputInfo))
	outputChannel := make(chan (clientState))
	timeoutChannel := make(chan (int))
	closeGameChannel := make(chan (int))
	newGameChannel := make(chan (bool))
	removePlayerChannel := make(chan [2]int)
	wordCheck, _ = sql.Open("sqlite3", "./words.db")
	// defer close(inputChannel)
	// defer close(outputChannel)
	// defer close(timeoutChannel)
	go server(inputChannel, timeoutChannel, outputChannel, newGameChannel, closeGameChannel, removePlayerChannel)
	game(inputChannel, timeoutChannel, outputChannel, newGameChannel, closeGameChannel, removePlayerChannel)
}
