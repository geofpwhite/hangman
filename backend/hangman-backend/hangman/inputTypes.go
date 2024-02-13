package hangman

import "fmt"

type input interface {
	ChangeStateAccordingToInput(outputChannel chan clientState)
}
type usernameInput struct {
	Username    string
	GameIndex   int
	PlayerIndex int
}

type newWordInput struct {
	NewWord     string
	GameIndex   int
	PlayerIndex int
}

type guessInput struct {
	Guess       string
	GameIndex   int
	PlayerIndex int
}

func (ui *usernameInput) ChangeStateAccordingToInput(outputChannel chan clientState) {
	fmt.Println("username change")
	fmt.Println(ui)
	fmt.Println(len(gStates))
	fmt.Println(len(gStates) <= ui.GameIndex)
	if len(gStates) <= ui.GameIndex {
		return
	}
	fmt.Println("username change")
	gState := gStates[ui.GameIndex]
	if len(gState.players) <= ui.PlayerIndex {
		return
	}
	fmt.Println("username change")
	gState.mut.Lock()
	(*gState).players[ui.PlayerIndex].username = ui.Username
	gState.mut.Unlock()
	outputChannel <- clientState{}

}
func (nwi *newWordInput) ChangeStateAccordingToInput(outputChannel chan clientState) {
	fmt.Println("word change")
	if len(gStates) <= nwi.GameIndex {
		return
	}
	gState := gStates[nwi.GameIndex]
	if len(gState.players) <= nwi.PlayerIndex {
		return
	}
	if (*gState).needNewWord && nwi.PlayerIndex == (*gState).curHostIndex {
		(*gState).newWord(nwi.NewWord, outputChannel)
	} else {
		outputChannel <- clientState{Warning: "you can't pick the word right now", PlayerIndex: nwi.PlayerIndex}
	}
}
func (gi *guessInput) ChangeStateAccordingToInput(outputChannel chan clientState) {
	fmt.Println("guess change")
	if len(gStates) <= gi.GameIndex {
		return
	}
	gState := gStates[gi.GameIndex]
	if len(gState.players) <= gi.PlayerIndex {
		return
	}

	if gi.PlayerIndex == (*gState).turn {
		// fmt.Println("guess")
		(*gState).guess(rune(gi.Guess[0]), outputChannel)
	} else {
		outputChannel <- clientState{Warning: "not your turn", PlayerIndex: gi.PlayerIndex}
	}
}
