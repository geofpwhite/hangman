package hangman

func test(inputChannel chan input, timeoutChannel chan int, outputChannel chan clientState, newGameChannel chan bool, closeGameChannel chan int, removePlayerChannel chan [2]int) {
	newGame()
}

func onePersonByThemself(inputChannel chan input, timeoutChannel chan int, outputChannel chan clientState, newGameChannel chan bool, closeGameChannel chan int, removePlayerChannel chan [2]int) {

}
