package hangman

/*
Interface implemented by user input objects to be accepted by game loop
*/
type input interface {
	ChangeStateAccordingToInput(chan clientState)
	GetGameHash() string
	GetPlayerIndex() int
}
type usernameInput struct {
	Username    string
	GameHash    string
	PlayerIndex int
}
type newWordInput struct {
	NewWord     string
	GameHash    string
	PlayerIndex int
}
type randomlyChooseWordInput struct {
	Guess       string
	GameHash    string
	PlayerIndex int
}
type guessInput struct {
	Guess       string
	GameHash    string
	PlayerIndex int
}
type chatInput struct {
	Message     string
	GameHash    string
	PlayerIndex int
}
type ruleChangeInput struct {
}

func (ui *usernameInput) GetGameHash() string {
	return ui.GameHash
}
func (ui *usernameInput) GetPlayerIndex() int {
	return ui.PlayerIndex
}
func (ui *usernameInput) ChangeStateAccordingToInput(outputChannel chan clientState) {
	if validateGameHashAndPlayerIndex(ui.GameHash, ui.PlayerIndex) {
		gState := gameHashes[ui.GameHash]
		gState.changeUsername(ui.PlayerIndex, ui.Username)
		outputChannel <- clientState{GameHash: gState.gameHash}
	}
}

func (nwi *newWordInput) GetGameHash() string {
	return nwi.GameHash
}
func (nwi *newWordInput) GetPlayerIndex() int {
	return nwi.PlayerIndex
}
func (nwi *newWordInput) ChangeStateAccordingToInput(outputChannel chan clientState) {
	if validateGameHashAndPlayerIndex(nwi.GameHash, nwi.PlayerIndex) {
		gState := gameHashes[nwi.GameHash]

		if gState.needNewWord && nwi.PlayerIndex == gState.curHostIndex {
			gState.newWord(nwi.NewWord)
			outputChannel <- clientState{GameHash: gState.gameHash}
		} else {
			outputChannel <- clientState{Warning: "you can't pick the word right now", PlayerIndex: nwi.PlayerIndex}
		}
	}
}
func (gi *guessInput) GetGameHash() string {
	return gi.GameHash
}
func (gi *guessInput) GetPlayerIndex() int {
	return gi.PlayerIndex
}

func (gi *guessInput) ChangeStateAccordingToInput(outputChannel chan clientState) {
	if validateGameHashAndPlayerIndex(gi.GameHash, gi.PlayerIndex) {
		gState := gameHashes[gi.GameHash]
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

func (ci *chatInput) GetGameHash() string {
	return ci.GameHash
}
func (ci *chatInput) GetPlayerIndex() int {
	return ci.PlayerIndex
}
func (ci *chatInput) ChangeStateAccordingToInput(outputChannel chan clientState) {
	if validateGameHashAndPlayerIndex(ci.GameHash, ci.PlayerIndex) {
		gState := gameHashes[ci.GameHash]
		gState.chat(ci.Message, ci.PlayerIndex)
		outputChannel <- clientState{GameHash: ci.GameHash}
	}
}

func (rcwi *randomlyChooseWordInput) GetGameHash() string {
	return rcwi.GameHash
}
func (rcwi *randomlyChooseWordInput) GetPlayerIndex() int {
	return rcwi.PlayerIndex
}
func (rcwi *randomlyChooseWordInput) ChangeStateAccordingToInput(outputChannel chan clientState) {
	if validateGameHashAndPlayerIndex(rcwi.GameHash, rcwi.PlayerIndex) {
		gState := gameHashes[rcwi.GameHash]
		gState.randomNewWord()
		outputChannel <- clientState{GameHash: rcwi.GameHash}
	}
}
