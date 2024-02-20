package hangman

import (
	// "fmt"
	"log"
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

func validatePlayerIndexAndGameIndex(gameIndex, playerIndex int) bool {
	if gameIndex >= len(gStates) || gameIndex < 0 {
		return false
	}
	gState := gStates[gameIndex]
	if playerIndex >= len(gState.players) || playerIndex < 0 {
		return false
	}
	return true
}

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
			log.Println("close game channel")
			if gameIndex < len(gStates) && gameIndex >= 0 {
				gStates[gameIndex].gameIndex = gameIndex
				gStates[gameIndex].closeGame()
				tickerInputChannels[gameIndex] <- inputInfo{GameIndex: -1}
				tickerInputChannels = slices.Delete(tickerInputChannels, gameIndex, gameIndex+1)
			}
		}
	}()
	for {
		select {
		case removePlayer := <-removePlayerChannel:
			log.Println("removePlayerChannel")
			if validatePlayerIndexAndGameIndex(removePlayer[0], removePlayer[1]) {
				playerIndex := removePlayer[1]
				gState := gStates[removePlayer[0]]
				if len(gState.players) <= 1 {
					closeGameChannel <- gState.gameIndex
				} else {
					go func() {
						gState.removePlayer(playerIndex)
						outputChannel <- clientState{GameIndex: gState.gameIndex}
					}()
				}
			}

		case <-newGameChannel:
			log.Println("newGameChannel")
			gState := newGame()
			newTickerInputChannel := make(chan (inputInfo))
			tickerInputChannels = append(tickerInputChannels, newTickerInputChannel)
			go (*gState).runTicker(tickerTimeoutChannel, newTickerInputChannel, closeGameChannel)

		case gameIndex := <-tickerTimeoutChannel:
			log.Println("tickertimeoutchannel")
			if gameIndex >= len(gStates) {
				continue
			}
			gState := gStates[gameIndex]
			if len((*gState).players) <= 1 {
				continue
			}
			go func() {
				timeoutChannel <- gState.handleTickerTimeout(timeoutChannel)
			}()

		case info := <-inputChannel:
			log.Println("input channel")
			if validatePlayerIndexAndGameIndex(info.GetGameIndex(), info.GetPlayerIndex()) {
				tickerInputChannels[info.GetGameIndex()] <- inputInfo{PlayerIndex: info.GetPlayerIndex()}
				log.Println(info)
				go info.ChangeStateAccordingToInput(outputChannel)
			}
		}
	}
}
