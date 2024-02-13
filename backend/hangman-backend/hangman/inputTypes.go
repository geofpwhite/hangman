package hangman

type input interface {
	ChangeStateAccordingToInput(outputChannel chan clientState)
	GetGameIndex() int
	GetPlayerIndex() int
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

type chatInput struct {
}

type ruleChangeInput struct {
}

func (ui *usernameInput) GetGameIndex() int {
	return ui.GameIndex
}
func (ui *usernameInput) GetPlayerIndex() int {
	return ui.PlayerIndex
}
func (nwi *newWordInput) GetGameIndex() int {
	return nwi.GameIndex
}
func (nwi *newWordInput) GetPlayerIndex() int {
	return nwi.PlayerIndex
}
func (gi *guessInput) GetGameIndex() int {
	return gi.GameIndex
}
func (gi *guessInput) GetPlayerIndex() int {
	return gi.PlayerIndex
}

func (ui *usernameInput) ChangeStateAccordingToInput(outputChannel chan clientState) {
	if len(gStates) <= ui.GameIndex {
		return
	}
	gState := gStates[ui.GameIndex]
	if len(gState.players) <= ui.PlayerIndex {
		return
	}
	gState.mut.Lock()
	(*gState).players[ui.PlayerIndex].username = ui.Username
	gState.mut.Unlock()
	outputChannel <- clientState{}

}
func (nwi *newWordInput) ChangeStateAccordingToInput(outputChannel chan clientState) {
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
