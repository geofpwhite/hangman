/*
test functions
*/
package hangman

import (
	"testing"
)

func TestAll(t *testing.T) {
	t.Run("11", TestOnePersonOneGame)
	t.Run("21", TestTwoPeopleOneGame)
	t.Run("52", TestFivePeopleTwoGames)
}

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
func TestTwoPeopleOneGame(t *testing.T) {
	gState := newGame()
	original := *gState
	gState.newPlayer(player{username: "player 1"})
	gState.newPlayer(player{username: "player 2"})
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

func TestFivePeopleTwoGames(t *testing.T) {
	// First game
	gState1 := newGame()
	original1 := *gState1
	gState1.newPlayer(player{username: "player 1"})
	gState1.newPlayer(player{username: "player 2"})
	gState1.newWord("phantom")
	gState1.guess('a')
	gState1.guess('a')
	gState1.guess('j')
	gState1.guess('p')
	gState1.guess('h')
	original1.guessed = "ajph"

	// Example assertion for the first game
	if gState1.winner != -1 {
		t.Errorf("Expected winner to be HOST_LOSES for game 1, got %d", gState1.winner)
	}
	if original1.guessed != gState1.guessed {
		t.Errorf("Expected guessed to be 'ajph' for game 1, got %s", gState1.guessed)
	}

	// Second game
	gState2 := newGame()
	original2 := *gState2
	gState2.newPlayer(player{username: "player 1"})
	gState2.newPlayer(player{username: "player 2"})
	gState2.newPlayer(player{username: "player 3"})
	gState2.newPlayer(player{username: "player 4"})
	gState2.newPlayer(player{username: "player 5"})
	gState2.newWord("example")
	gState2.guess('e')
	gState2.guess('e')
	gState2.guess('j')
	gState2.guess('p')
	gState2.guess('l')
	original2.guessed = "ejpl"

	// Example assertion for the second game
	if gState2.winner != -1 {
		t.Errorf("Expected winner to be HOST_LOSES for game 2, got %d", gState2.winner)
	}
	if original2.guessed != gState2.guessed {
		t.Errorf("Expected guessed to be 'ejpl' for game 2, got %s", gState2.guessed)
	}
}
