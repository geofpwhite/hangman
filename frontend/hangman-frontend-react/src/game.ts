import { setHashCookie } from "./App"
import { chatMessage } from "./chat"

export interface GameState {
  needNewWord: boolean
  players: Array<string>
  playerIndex: number
  turn: number
  host: number
  revealedWord: string
  lettersGuessed: string
  guessesLeft: number
  winner: number
  gameIndex: number
  chatLogs: Array<chatMessage>
  hash: string
}
export class Game {
  state: GameState = {

    needNewWord: false,
    playerIndex: -1,
    players: [],
    turn: -1,
    host: -1,
    revealedWord: "",
    lettersGuessed: "",
    guessesLeft: -1,
    winner: -1,
    gameIndex: -1,
    chatLogs: [],
    hash: "",

  }


  fromGameState(gs: GameState) {
    if (gs.hash !== "") {
      this.state.hash = gs.hash
      setHashCookie(gs.hash)
    }
    if (gs.needNewWord != null) {
      this.state.needNewWord = gs.needNewWord
    }
    if (gs.players.length > 0) {
      this.state.players = gs.players
    }
    if (gs.turn !== -1) {
      this.state.turn = gs.turn
    }
    if (gs.host !== -1) {
      this.state.host = gs.host
    }
    if (gs.revealedWord !== "") {
      this.state.revealedWord = gs.revealedWord
    }
    if (gs.lettersGuessed !== null) {
      this.state.lettersGuessed = gs.lettersGuessed
    }
    if (gs.guessesLeft !== -1) {
      this.state.guessesLeft = gs.guessesLeft
    }
    if (gs.playerIndex !== -1) {
      this.state.playerIndex = gs.playerIndex
    }
    if (gs.gameIndex !== -1) {
      this.state.gameIndex = gs.gameIndex
    }
    if (gs.chatLogs !== null) {
      this.state.chatLogs = gs.chatLogs
    } else {
      this.state.chatLogs = []
    }
    this.state.winner = gs.winner
  }
}
