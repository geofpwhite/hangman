/*
test functions
*/
package hangman

import (
	"fmt"
	"testing"
)

func TestOnePersonOneGame(t *testing.T) {
	gState := newGame()
	original := *gState
	gState.newPlayer(player{username: "player 1"})
	gState.newWord("phantom")
	gState.guess('a')
	gState.guess('a')
	gState.guess('j')
	gState.guess('p')
	gState.guess('h')
	original.guessed = "ajph"

	// Example assertion
	if gState.winner != -1 {
		t.Errorf("Expected winner to be HOST_LOSES, got %d", gState.winner)
	}
	if original.guessed != gState.guessed {
		t.Errorf("Expected guessed to be 'ajph', got %s", gState.guessed)
	}
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
