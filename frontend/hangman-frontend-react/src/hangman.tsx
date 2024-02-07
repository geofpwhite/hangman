import React, { useEffect, useState } from 'react';
import "./hangman.css";
import { Game, GameState } from './game';
import GAME_OVER from "./game_over.png";
import LEFT1 from "./1left.png"
import LEFT2 from "./2left.png"
import LEFT3 from "./3left.png"
import LEFT4 from "./4left.png"
import LEFT5 from "./5left.png"
import LEFT6 from "./6left.png"
import LEFT7 from "./7left.png"
import LEFT8 from "./8left.png"
import LEFT9 from "./9left.png"
import GRATEFUL from "./grateful.jpeg"



const HOST_WINS = 0
const HOST_LOSES = 1

const guessesLeftImages = [GAME_OVER, LEFT1, LEFT2, LEFT3, LEFT4, LEFT5, LEFT6, LEFT7, LEFT8, LEFT9]

interface HangmanComponentProps {
  gameIndex: number
}

const HangmanComponent: React.FC<HangmanComponentProps> = ({ gameIndex }) => {
  const [webSocket, setWebSocket] = useState<WebSocket | null>(null);
  const [gameState, setGameState] = useState<GameState>();
  const [usernameInputValue, setInputValue] = useState('');
  const [newWordInputValue, setInputValue2] = useState('');
  const letters: string = "abcdefghijklmnopqrstuvwxyz";
  var game = new Game()


  const drawHangMan = () => {
    if (!gameState) {
      return
    }
    if (gameState.guessesLeft >= 10) {
      return <div style={{ width: "300px", height: "100px" }} />
    } else if (gameState?.guessesLeft < 10) {
      if (gameState.needNewWord) {
        if (gameState.winner === HOST_WINS) {

          return <img src={GAME_OVER} style={{ width: "300px", height: "100px" }} alt="" />
        } else {
          return <img src={GRATEFUL} style={{ width: "300px", height: "100px" }} alt="" />
        }
      }
      return (
        <img src={guessesLeftImages[gameState?.guessesLeft]} style={{ width: "300px", height: "100px" }} alt="" />
      )
    }

  }

  useEffect(() => {
    const ws = new WebSocket('ws://localhost:8000/ws/' + gameIndex);

    ws.onopen = () => {
      console.log('WebSocket connection opened');
    };

    ws.onmessage = (event) => {
      try {
        let obj = JSON.parse(event.data);
        setGameState(() => {
          // Update the state based on the previous state
          game.fromGameState(obj);
          return { ...game.state };
        });
      } catch (error) {
        console.log('Error parsing WebSocket message:', error);
      }
    };

    ws.onclose = () => {
      console.log('WebSocket connection closed');
    };

    setWebSocket(ws);

    return () => {
      ws.close();
    };
  }, []); // Empty dependency array ensures this effect runs only once


  const sendGuess = (letter: string) => {
    webSocket?.send("g:" + letter);
  };

  const sendNewWord = () => {
    let c = newWordInputValue.toLowerCase()
    for (let i = 0; i < newWordInputValue.length; i++) {
      if (!letters.includes(c[i])) {
        //send error, letters only
        return
      }
    }
    webSocket?.send("w:" + c);
  };

  const changeUsername = () => {
    webSocket?.send("u:" + usernameInputValue);
    setInputValue('');
    setWantsToChangeUsername(false)
  };



  const handleChange2 = (event: React.ChangeEvent<HTMLInputElement>) => {
    setInputValue2(event.target.value);
  };

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setInputValue(event.target.value);
  };

  const determineColor = (letter: string): string => {
    if (!gameState || gameState.needNewWord) {
      return 'black';
    }

    if (gameState.lettersGuessed.includes(letter)) {
      if (gameState.revealedWord.includes(letter)) {
        return 'green';
      } else return 'red';
    } else return 'black';
  };

  const PlayerNames = () => {
    if (!gameState) {
      return
    } else {
      return (
        <div>
          your username is {gameState.players[gameState.playerIndex]}, it is {gameState.players[gameState.turn]}'s turn
          {
            gameState.players.map((value: string, id: number) => (
              <div style={{ color: id === gameState.turn ? 'green' : id === gameState.host ? 'red' : 'black' }}>
                {value}
              </div>
            ))}
        </div>
      )
    }
  }
  console.log(gameState)

  const NewWordInputBox = () => {
    if (!gameState || !gameState.needNewWord) {
      return (<div></div>)
    }
    if (gameState?.needNewWord && gameState?.host === gameState?.playerIndex) {

      return (
        <div>
          <div className="new-word">
            <div>
              <label htmlFor="newWord">We need a new word </label>
              <input
                type="text"
                id="myInput"
                value={newWordInputValue}
                onChange={handleChange2}
                placeholder="Type here..."
              />
              <button type="button" onClick={sendNewWord}>Submit</button>
            </div>
          </div>

        </div>
      )
    }
  }


  const determineHostAndTurnDisplay = () => {
    let host: boolean = false
    let turn: boolean = false
    if (gameState?.host === gameState?.playerIndex) {
      host = true
    }
    if (!host) {
      if (gameState?.turn === gameState?.playerIndex) {
        turn = true
      }
    }

    let turnText
    if (turn) {
      turnText = (
        <div>
          <h2>
            It's your turn
          </h2>
        </div>
      )
    } else {
      turnText = (
        <div>
          <h2>
            Not your turn
          </h2>
        </div>
      )
    }
    if (host) {
      turnText = (
        <div>
          <h2>
            You're the host
          </h2>
        </div>
      )
    }
    return turnText
  }


  const [wantsToChangeUsername, setWantsToChangeUsername] = useState<boolean>(false);

  const usernameInputBox = () => {
    if (wantsToChangeUsername) {
      return (
        <div className="change-username">
          <div>
            <label htmlFor="myInput">Username: </label>
            <input
              type="text"
              id="myInput"
              value={usernameInputValue}
              onChange={handleChange}
              placeholder="Type here..."
            />
            <button onClick={changeUsername}>Submit</button>
          </div>
        </div>
      )
    } else {
      return (
        <div>
          <button type="button" onClick={() => { setWantsToChangeUsername(true) }}>Change Username</button>
        </div>
      )
    }
  }

  const splitWord = () => {
    var str = ""
    gameState?.revealedWord.split("").forEach((value, index) => {
      str += " " + value
    })
    return str
  }

  const determineGuessOrCongrats = () => {

    if (gameState?.needNewWord) {

      if (gameState.winner === HOST_WINS && gameState.playerIndex === gameState.host - 1) {
        return "You Won"
      } else if (gameState.winner === HOST_WINS) {
        return "You Lost"
      } else if (gameState.winner === HOST_LOSES && gameState.playerIndex === gameState.host - 1) {
        return "You Lost"
      } else if (gameState.winner === HOST_LOSES) {
        return "You Won"
      }
    } else {

      return (gameState?.guessesLeft + " Guesses Left")

    }
  }



  return (
    <div>
      {
        drawHangMan()
      }
      {
        determineHostAndTurnDisplay()
      }
      <div className="theWord">
        <h1>
          {
            splitWord()
          }
        </h1>
      </div>

      <div className="guessesLeft">
        <h1>
          {
            determineGuessOrCongrats()
          }
        </h1>
      </div>

      <div className="lettersGuessed">
        <p>
          Letters Guessed
        </p>
        <p id="letters-guessed">
          {gameState?.lettersGuessed}
        </p>
      </div>
      {
        PlayerNames()
      }

      {
        usernameInputBox()
      }

      {
        gameState?.needNewWord ? NewWordInputBox() : (<div />)
      }
      <div className="letter-grid">
        {letters.split("").map((value: string, id: number) => (
          // <div className="letter-item" style={{ color: determineColor(value), backgroundColor: "transparent" }}>
          <button type="button" key={id} className="letter-button" onClick={() => sendGuess(value)} style={{ color: determineColor(value) }}>
            {value}
          </button>
          // </div>
        ))}
      </div>
    </div>
  );
};

export default HangmanComponent;
