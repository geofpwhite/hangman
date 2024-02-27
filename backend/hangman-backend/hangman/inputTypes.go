package hangman

/*
Interface implemented by user input objects to be accepted by game loop
*/
type input interface {
	ChangeStateAccordingToInput(chan clientState)
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
	Message     string
	GameIndex   int
	PlayerIndex int
}
type ruleChangeInput struct {
}

func (ui *usernameInput) GetGameIndex() int {
	return ui.GameIndex
}
func (ui *usernameInput) GetPlayerIndex() int {
	return ui.PlayerIndex
}
func (ui *usernameInput) ChangeStateAccordingToInput(outputChannel chan clientState) {
	if validateGameIndexAndPlayerIndex(ui.GameIndex, ui.PlayerIndex) {
		gState := gStates[ui.GameIndex]
		gState.changeUsername(ui.PlayerIndex, ui.Username)
		outputChannel <- clientState{GameIndex: gState.gameIndex}
	}
}

func (nwi *newWordInput) GetGameIndex() int {
	return nwi.GameIndex
}
func (nwi *newWordInput) GetPlayerIndex() int {
	return nwi.PlayerIndex
}
func (nwi *newWordInput) ChangeStateAccordingToInput(outputChannel chan clientState) {
	if validateGameIndexAndPlayerIndex(nwi.GameIndex, nwi.PlayerIndex) {
		gState := gStates[nwi.GameIndex]

		if gState.needNewWord && nwi.PlayerIndex == gState.curHostIndex {
			gState.newWord(nwi.NewWord)
			outputChannel <- clientState{GameIndex: gState.gameIndex}
		} else {
			outputChannel <- clientState{Warning: "you can't pick the word right now", PlayerIndex: nwi.PlayerIndex}
		}
	}
}
func (gi *guessInput) GetGameIndex() int {
	return gi.GameIndex
}
func (gi *guessInput) GetPlayerIndex() int {
	return gi.PlayerIndex
}

func (gi *guessInput) ChangeStateAccordingToInput(outputChannel chan clientState) {
	if validateGameIndexAndPlayerIndex(gi.GameIndex, gi.PlayerIndex) {
		gState := gStates[gi.GameIndex]
		if gi.PlayerIndex == gState.turn {
			// fmt.Println("guess")
			output, changedParts := gState.guess(rune(gi.Guess[0]))
			if output {
				outputChannel <- changedParts
			}
		} else {
			outputChannel <- clientState{Warning: "not your turn", PlayerIndex: gi.PlayerIndex}
		}
	}
}

func (ci *chatInput) GetGameIndex() int {
	return ci.GameIndex
}
func (ci *chatInput) GetPlayerIndex() int {
	return ci.PlayerIndex
}
func (ci *chatInput) ChangeStateAccordingToInput(outputChannel chan clientState) {
	if validateGameIndexAndPlayerIndex(ci.GameIndex, ci.PlayerIndex) {
		gState := gStates[ci.GameIndex]
		gState.chat(ci.Message, ci.PlayerIndex)
		outputChannel <- clientState{GameIndex: ci.GameIndex}
	}
}
