package hangman

import (
	// "fmt"
	"slices"

	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
)

const HOST_WINS = 1
const HOST_LOSES = 2

type player struct {
	username   string
	connection *websocket.Conn
}

// var gState serverState = serverState{}
var gStates []*gameState = []*gameState{}

func game(
	inputChannel chan input,
	timeoutChannel chan int,
	outputChannel chan clientState,
	newGameChannel chan bool,
	closeGameChannel chan int,
	removePlayerChannel chan [2]int,
) {

	tickerInputChannels := []chan inputInfo{}
	tickerTimeoutChannel := make(chan (int))
	{
		tickerInputChannel := make(chan (inputInfo))
		tickerInputChannels = append(tickerInputChannels, tickerInputChannel)
	}
	go gStates[0].runTicker(tickerTimeoutChannel, tickerInputChannels[0], closeGameChannel)
	go func() {
		for gameIndex := range closeGameChannel {
			println("close game channel")
			if gameIndex < len(gStates) {
				gStates[gameIndex].gameIndex = gameIndex
				gStates[gameIndex].closeGame()
				tickerInputChannels[gameIndex] <- inputInfo{GameIndex: -1}
				close(tickerInputChannels[gameIndex])
				tickerInputChannels = slices.Delete(tickerInputChannels, gameIndex, gameIndex+1)
			}
		}
	}()
	for {
		select {
		case removePlayer := <-removePlayerChannel:
			println("removePlayerChannel")
			if removePlayer[0] >= len(gStates) {
				continue
			}
			gState := gStates[removePlayer[0]]
			if removePlayer[1] >= len(gState.players) || removePlayer[0] < 0 || removePlayer[1] < 0 {
				continue
			}
			playerIndex := removePlayer[1]
			gState.removePlayer(playerIndex, tickerInputChannels, outputChannel, closeGameChannel)

		case <-newGameChannel:
			println("newGameChannel")
			gState := newGame()
			newTickerInputChannel := make(chan (inputInfo))
			tickerInputChannels = append(tickerInputChannels, newTickerInputChannel)
			go (*gState).runTicker(tickerTimeoutChannel, newTickerInputChannel, closeGameChannel)

		case gameIndex := <-tickerTimeoutChannel:
			println("tickertimeoutchannel")
			if gameIndex >= len(gStates) {
				continue
			}
			gState := gStates[gameIndex]
			//timed out, move to the next player
			if len((*gState).players) <= 1 {
				closeGameChannel <- gameIndex
				continue
			}
			gState.mut.Lock()
			if gState.needNewWord {
				gState.curHostIndex = (gState.curHostIndex + 1) % len(gState.players)
				gState.turn = gState.curHostIndex
			} else {

				(*gState).turn = ((*gState).turn + 1) % len((*gState).players)
				if (*gState).curHostIndex == (*gState).turn {
					(*gState).turn = ((*gState).turn + 1) % len((*gState).players)
				}
			}
			gState.mut.Unlock()
			timeoutChannel <- gameIndex //for the websocket to update everybody

		case info := <-inputChannel:
			println("input channel")
			info.ChangeStateAccordingToInput(outputChannel)
		}

	}
}
