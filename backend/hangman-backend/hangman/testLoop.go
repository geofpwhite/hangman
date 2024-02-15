package hangman

import "fmt"

/*
test function
*/
func test() {
	gState := newGame()
	gState.newPlayer(player{username: "player 1"})
	gState.newWord("phantom")
	gState.guess('a')
	gState.newPlayer(player{username: "player 2"})
	fmt.Println(gState)
}

func onePersonByThemself() {
	gState := newGame()
	gState.newPlayer(player{username: "player 1"})
	gState.newWord("phantom")
	gState.guess('a')
	gState.guess('a')
	gState.guess('j')
	gState.guess('p')
	gState.guess('h')
	fmt.Println(gState)
}
func twoPlayers() {
	gState := newGame()
	gState.newPlayer(player{username: "player 1"})
	gState.newWord("phantom")
	gState.guess('a')
	gState.guess('a')
	gState.guess('j')
	gState.guess('p')
	gState.guess('h')
	fmt.Println(gState)
}
func twoGamesMultiplePlayers() {

}
