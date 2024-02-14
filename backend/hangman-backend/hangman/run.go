package hangman

import (
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type inputInfo struct {
	Username    string `json:"username"`
	Guess       string `json:"guess"`
	Word        string `json:"word"`
	PlayerIndex int    `json:"playerIndex"`
	GameIndex   int    `json:"gameIndex"`
}

/*  */
type clientState struct {
	Players        []string  `json:"players"`
	Turn           int       `json:"turn"`
	Host           int       `json:"host"`
	RevealedWord   string    `json:"revealedWord"`
	GuessesLeft    int       `json:"guessesLeft"`
	LettersGuessed string    `json:"lettersGuessed"`
	NeedNewWord    bool      `json:"needNewWord"`
	Warning        string    `json:"warning"`
	PlayerIndex    int       `json:"playerIndex"` // changes for each connection that the update state object is sent to
	Winner         int       `json:"winner"`
	GameIndex      int       `json:"gameIndex"`
	ChatLogs       []chatLog `json:"chatLogs"`
}

func Run() {
	newGame()
	inputChannel := make(chan (input))
	outputChannel := make(chan (clientState))
	timeoutChannel := make(chan (int))
	closeGameChannel := make(chan (int))
	newGameChannel := make(chan (bool))
	removePlayerChannel := make(chan [2]int)
	os.Remove("app.log")
	file, _ := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	log.SetOutput(file)
	// defer close(inputChannel)
	// defer close(outputChannel)
	// defer close(timeoutChannel)
	go game(inputChannel, timeoutChannel, outputChannel, newGameChannel, closeGameChannel, removePlayerChannel)
	server(inputChannel, timeoutChannel, outputChannel, newGameChannel, closeGameChannel, removePlayerChannel)
}
func testRun() {
	newGame()
	inputChannel := make(chan (input))
	outputChannel := make(chan (clientState))
	timeoutChannel := make(chan (int))
	closeGameChannel := make(chan (int))
	newGameChannel := make(chan (bool))
	removePlayerChannel := make(chan [2]int)
	// defer close(inputChannel)
	// defer close(outputChannel)
	// defer close(timeoutChannel)
	go game(inputChannel, timeoutChannel, outputChannel, newGameChannel, closeGameChannel, removePlayerChannel)
	test(inputChannel, timeoutChannel, outputChannel, newGameChannel, closeGameChannel, removePlayerChannel)
}

//func test(inputChannel, timeoutChannel, outputChannel, newGameChannel, closeGameChannel, removePlayerChannel)
