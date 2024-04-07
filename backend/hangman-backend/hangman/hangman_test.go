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
	t.Run("add-and-remove", TestAddingAndRemovingPlayers)
	t.Run("word-change-host-loses", TestSuccessfulGuessAndHostChange)
	t.Run("word-change-host-wins", TestHostWinsAndHostChange)
}

func TestAddingAndRemovingPlayers(t *testing.T) {
	gState := newGame()
	originalNumPlayers := len(gState.players)
	// Test adding players
	gState.newPlayer(player{username: "player 1"})
	gState.newPlayer(player{username: "player 2"})
	if len(gState.players) != originalNumPlayers+2 {
		t.Errorf("Expected %d players, got %d", originalNumPlayers+2, len(gState.players))
	}
	gState.newPlayer(player{username: "player 2"})
	gState.newPlayer(player{username: "player 3"})
	gState.removePlayer(1)

	// Test removing players
	gState.removePlayer(0)
	if len(gState.players) != originalNumPlayers+2 {
		t.Errorf("Expected %d players after removal, got %d", originalNumPlayers+1, len(gState.players))
	}
}

func TestSuccessfulGuessAndHostChange(t *testing.T) {
	gState := newGame()
	gState.newPlayer(player{username: "player 1"})
	gState.newPlayer(player{username: "player 2"})
	gState.newWord("example")
	gState.guess('e') // Successful guess
	gState.guess('x') // Successful guess
	gState.guess('a') // Successful guess
	gState.guess('m') // Successful guess
	gState.guess('p') // Successful guess
	gState.guess('l') // Successful guess
	if gState.curHostIndex == 0 {
		t.Errorf("Expected host to change after a successful guess, but it didn't.")
	}
	if !gState.needNewWord {
		t.Errorf("Expected needNewWord to be true, got false")
	}
}

func TestHostWinsAndHostChange(t *testing.T) {
	gState := newGame()
	gState.newPlayer(player{username: "player 1"})
	gState.newPlayer(player{username: "player 2"})
	gState.newWord("example")
	// Simulate host winning
	gState.guess('z')
	gState.guess('y')
	gState.guess('g')
	gState.guess('u')
	gState.guess('v')
	gState.guess('r')
	if gState.curHostIndex != 1 {
		t.Errorf("Expected host to be 1 after host wins, but got %d.", gState.curHostIndex)
	}
	if !gState.needNewWord {
		t.Errorf("Expected needNewWord to be true, got false")
	}
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
		t.Errorf("Expected winner to be -1, got %d", gState.winner)
	}
	if original.guessed != gState.guessed {
		t.Errorf("Expected guessed to be 'ajph', got %s", gState.guessed)
	}
}

func TestFivePeopleTwoGames(t *testing.T) {
	// First game
	gState1 := newGame()
	gState2 := newGame()
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

	// Example assertion for the first game
	if gState1.winner != -1 {
		t.Errorf("Expected winner to be HOST_LOSES for game 1, got %d", gState1.winner)
	}
	if original1.guessed != gState1.guessed {
		t.Errorf("Expected guessed to be 'ajph' for game 1, got %s", gState1.guessed)
	}

	if gState2.winner != -1 {
		t.Errorf("Expected winner to be HOST_LOSES for game 2, got %d", gState2.winner)
	}
	if original2.guessed != gState2.guessed {
		t.Errorf("Expected guessed to be 'ejpl' for game 2, got %s", gState2.guessed)
	}
	if gState2.curHostIndex != 0 {
		t.Errorf("Expected host to be player 1, got %s", gState2.players[gState2.curHostIndex].username)
	}
}
