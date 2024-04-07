package hangman

import (
	// "fmt"

	"log"
	"time"

	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
)

const HOST_WINS = 1
const HOST_LOSES = 2

type player struct {
	username   string
	connection *websocket.Conn
	hash       string
}

// var gState serverState = serverState{}
var gameHashes map[string]*gameState = map[string]*gameState{}

func validateGameHashAndPlayerIndex(hash string, playerIndex int) bool {
	if gameHashes[hash] == nil {
		return false
	}
	gState := gameHashes[hash]
	if playerIndex >= len(gState.players) || playerIndex < 0 {
		return false
	}
	return true
}
func cleanupHashFunction(closeGameHashChannel chan string) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		for i := range gameHashes {
			if len(gameHashes[i].players) == 0 || gameHashes[i] == nil || gameHashes[i].consecutiveTimeouts >= len(gameHashes[i].players) {
				closeGameHashChannel <- i
			}
		}
	}

}

type playerToRemove struct {
	gameHash    string
	playerIndex int
}

func game(
	inputChannel chan input,
	timeoutChannel chan string,
	outputChannel chan clientState,
	newGameChannel chan bool,
	closeGameChannel chan string,
	removePlayerChannel chan playerToRemove,
	tickerInputChannels map[string]chan inputInfo,
	tickerTimeoutChannel chan string,
	emptyGameHash string,
) {

	{
		tickerInputChannel := make(chan (inputInfo))
		tickerInputChannels[emptyGameHash] = tickerInputChannel
	}
	go cleanupHashFunction(closeGameChannel)
	go gameHashes[emptyGameHash].runTicker(tickerTimeoutChannel, tickerInputChannels[emptyGameHash], closeGameChannel)
	go func() {
		for gameHash := range closeGameChannel {
			log.Println("close game channel")
			gameHashes[gameHash].closeGame()
			tickerInputChannels[gameHash] <- inputInfo{GameHash: ""}
			delete(tickerInputChannels, gameHash)
		}
	}()
	for {
		select {
		case removePlayer := <-removePlayerChannel:
			log.Println("removePlayerChannel")
			if validateGameHashAndPlayerIndex(removePlayer.gameHash, removePlayer.playerIndex) {
				playerIndex := removePlayer.playerIndex
				gState := gameHashes[removePlayer.gameHash]
				if len(gState.players) <= 1 {
					gState.removePlayer(playerIndex)
					closeGameChannel <- gState.gameHash
				} else {
					go func() {
						gState.removePlayer(playerIndex)
						if len(gState.players) == 0 {
							closeGameChannel <- gState.gameHash
						}
						outputChannel <- clientState{GameHash: gState.gameHash}
					}()
				}
			}

		case <-newGameChannel:
			log.Println("newGameChannel")
			gState := newGame()
			newTickerInputChannel := make(chan (inputInfo))
			tickerInputChannels[gState.gameHash] = newTickerInputChannel
			go (*gState).runTicker(tickerTimeoutChannel, newTickerInputChannel, closeGameChannel)

		case gameHash := <-tickerTimeoutChannel:
			log.Println("tickertimeoutchannel")

			gState := gameHashes[gameHash]
			if len((gState).players) <= 1 {
				continue
			}
			go func() {
				timeoutChannel <- gState.handleTickerTimeout()
			}()

		case info := <-inputChannel:
			log.Println("input channel")
			tickerInputChannels[info.GetGameHash()] <- inputInfo{PlayerIndex: info.GetPlayerIndex()}
			log.Println(info)
			go info.ChangeStateAccordingToInput(outputChannel)
		}
	}
}
