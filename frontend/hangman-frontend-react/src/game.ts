
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
        winner: -1
    }

    fromGameState(gs: GameState) {
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
        this.state.winner = gs.winner
    }
}
