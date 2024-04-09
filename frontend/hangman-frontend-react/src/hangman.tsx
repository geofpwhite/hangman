import React, { useEffect, useState } from 'react';
import axios from 'axios';
import "./hangman.css";
import { Game, GameState } from './game';
import GAME_OVER from "./game_over.png";
import GRATEFUL from "./grateful.jpeg";
import Chat from './chat';
import { setGameHashCookie, setHashCookie } from './App';


const HOST_WINS = 1
const HOST_LOSES = 2
var game = new Game()

interface HangmanComponentProps {
  gameHash: string
  reconnect: boolean
  hash: string
  reset: () => void
}


const HangmanComponent: React.FC<HangmanComponentProps> = ({ gameHash, reconnect, hash, reset }) => {
  const [webSocket, setWebSocket] = useState<WebSocket | null>(null);
  const [gameState, setGameState] = useState<GameState>();
  const [openChat, setOpenChat] = useState<boolean>(false);
  const [usernameInputValue, setInputValue] = useState('');
  const [newWordInputValue, setInputValue2] = useState('');
  const [wantsToChangeUsername, setWantsToChangeUsername] = useState<boolean>(false);
  const [newHash, setHash] = useState<string>('')
  const alphabet: string = "abcdefghijklmnopqrstuvwxyz";
  console.log(gameHash, reconnect, hash)


  const _url = "http://localhost:8080"
  // const _url = "https://hangman-backend-geoffrey.com"
  const stickFigure = (
    gameState ?
      gameState.winner === HOST_LOSES ?
        <img src={GRATEFUL} alt="" className="grateful-man" /> :
        <div className="stick-figure">
          <div className="gallows-top" style={{ backgroundColor: gameState!.guessesLeft < 8 ? "#333" : "white" }} />
          <div className="gallows-vertical" style={{ backgroundColor: gameState!.guessesLeft < 9 ? "#333" : "white" }} />
          <div className="rope" style={{ backgroundColor: gameState!.guessesLeft < 7 ? "#333" : "white" }} />
          <div className="head" style={{ border: gameState!.guessesLeft < 6 ? "1px solid #333" : "1px solid white" }} />
          <div className="body" style={{ backgroundColor: gameState!.guessesLeft < 5 ? "#333" : "white" }} />
          <div className="right-leg" style={{ backgroundColor: gameState!.guessesLeft < 4 ? "#333" : "white" }} />
          <div className="left-leg" style={{ backgroundColor: gameState!.guessesLeft < 3 ? "#333" : "white" }} />
          <div className="right-arm" style={{ backgroundColor: gameState!.guessesLeft < 2 ? "#333" : "white" }} />
          <div className="left-arm" style={{ backgroundColor: gameState.winner !== -1 ? "#333" : "white" }} />
          <div className="gallows-floor" style={{ backgroundColor: gameState!.guessesLeft < 10 ? "#333" : "white" }} />
        </div> : null
  )



  // const _url = "https://hangman-backend-geoffrey.com"
  // const _url = "http://localhost:8080"
  const exitGameButton = () => {
    console.log(hash)
    return (
      <div style={{ position: "absolute" }}>
        <button onClick={() => {
          exitGame(gameHash)
        }}>Exit Game</button>
      </div>
    )
  }
  const exitGame = (gameHash: string) => {
    console.log("hash  " + newHash)
    let nhash = newHash
    if (newHash === '') {
      nhash = hash
    }
    setHashCookie("")
    setGameHashCookie("")
    axios.get(_url + '/exit_game/' + nhash + '/' + gameHash,).then((response) => {
      console.log("response\n" + response)
      reset()
    })
  }

  useEffect(() => {
    var ws: WebSocket
    if (reconnect) {
      ws = new WebSocket('ws://localhost:8080/reconnect/' + hash + '/' + gameHash)
    } else {
      ws = new WebSocket('ws://localhost:8080/ws/' + gameHash)
    }

    ws.onopen = () => {
      console.log('WebSocket connection opened');
    };

    ws.onmessage = (event) => {
      try {
        let obj = JSON.parse(event.data);
        setGameState(() => {
          // Update the state based on the previous state
          console.log(obj + "          obj")
          game.fromGameState(obj);
          if (obj.hash !== '' && !reconnect) {
            setHash(obj.hash)
          }
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

  const chats = () => {
    if (gameState) {
      return (
        <Chat chats={gameState.chatLogs} sendMessage={sendChat} players={gameState.players} playerIndex={gameState.playerIndex} openChat={openChat}></Chat>
      )
    } else {
      return (
        <div />
      )
    }
  }

  const drawHangMan = () => {
    return stickFigure
  }


  const sendChat = (message: string) => {
    webSocket?.send("c:" + message)
  }

  const sendGuess = (letter: string) => {
    webSocket?.send("g:" + letter);
  };

  const sendRandomNewWordRequest = () => {
    webSocket?.send("r:")
  }

  const sendNewWord = () => {
    let c = newWordInputValue.toLowerCase()
    for (let i = 0; i < newWordInputValue.length; i++) {
      if (!alphabet.includes(c[i])) {
        //send error, letters only
        return
      }
    }
    webSocket?.send("w:" + c);
  };

  const changeUsername = () => {
    if (usernameInputValue.length <= 20) {
      webSocket?.send("u:" + usernameInputValue);
    }
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
        <div className="wrapper">
          your username is {gameState.players[gameState.playerIndex]}, it is {gameState.players[gameState.turn]}'s turn
          {
            gameState.players.map((value: string, id: number) => (
              <div key={id} style={{ color: id === gameState.turn ? 'green' : id === gameState.host ? 'red' : 'black' }}>
                {gameState.host === id ? value + "<- HOST" : value}
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
              <label htmlFor="myInput">
                We need a new word </label>
              <input
                type="text"
                id="myInput"
                value={newWordInputValue}
                onChange={handleChange2}
                placeholder="Type here..."
              />
              <button id="send-new-word" type="button" onClick={sendNewWord}>Submit</button>
            </div>
            <button id="random-new-word" type="button" onClick={sendRandomNewWordRequest}>Random word</button>
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
    if (gameState?.needNewWord) {
      if (!host) {
        turnText = (
          <div>
            <h2>
              waiting for the host
            </h2>
          </div>
        )
      }

    } else if (turn) {
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



  const usernameInputBox = () => {
    if (wantsToChangeUsername) {
      return (
        <div className="change-username">
          <div>
            <div>
              <label htmlFor="myInput">Username: </label>
            </div>
            <input
              type="text"
              id="myInput"
              value={usernameInputValue}
              onChange={handleChange}
              placeholder="Type here..."
            />
            <button id="change-username" onClick={changeUsername}>Submit</button>
          </div>
        </div>
      )
    } else {
      return (
        <div>
          <button id="wants-to-change-username" type="button" onClick={() => { setWantsToChangeUsername(true) }}>Change Username</button>
        </div>
      )
    }
  }

  const splitWord = () => {
    var str = ""
    gameState?.revealedWord.split("").forEach((value) => {
      str += " " + value
    })
    return str
  }

  const determineGuessOrCongrats = () => {

    if (gameState?.needNewWord) {
      if (gameState.winner === HOST_LOSES) {
        return gameState.players[(gameState.host + gameState.players.length - 1) % gameState.players.length] + "'s word was guessed "
      } else if (gameState.winner === HOST_WINS) {
        let previousHost = gameState.host
        if (previousHost === 0) {
          previousHost = gameState.players.length - 1
        } else {
          previousHost--
        }
        return gameState.players[previousHost] + " won"
      }

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

  const letterGrid = () => {
    if (gameState?.needNewWord) {
      return (<div />)
    }
    return (
      <div className="letter-grid">
        {alphabet.split("").map((value: string, id: number) => {
          if (!(gameState?.lettersGuessed.includes(value))) {
            return (
              <button id={"letter-" + value} type="button" key={id} className="letter-button" onClick={() => sendGuess(value)} >
                {value}
              </button>
            )
          } else {
            return (<div />)
          }
        })}
      </div>
    )
  }

  const lettersGuessed = () => {
    return (
      <p>
        {
          gameState?.lettersGuessed.split("").map((letter: string, _: number) => {
            return (
              <span id="letters-guessed" style={{ color: determineColor(letter) }}>
                {letter}
              </span>
            )
          })
        }
      </p>
    )
  }

  return (
    <div>
      {
        exitGameButton()
      }
      {
        drawHangMan()
      }
      {
        determineHostAndTurnDisplay()
      }
      <div className="the-word">
        <h1>
          {
            splitWord()
          }
        </h1>
      </div>

      <div className="guesses-left">
        <h1>
          {
            determineGuessOrCongrats()
          }
        </h1>
      </div>

      <div className="letters-guessed">
        {
          lettersGuessed()
        }
      </div>
      <div className="player-names">
        {
          PlayerNames()
        }
      </div>
      <div className="username-input-box">
        {
          usernameInputBox()
        }
      </div>
      <div className="new-word-input-box">
        {
          gameState?.needNewWord ? NewWordInputBox() : (<div />)
        }
      </div>
      {
        letterGrid()
      }
      <button id="open-chat" type="button"
        onClick={() => { setOpenChat(!openChat) }}
      >Toggle Chat
      </button>
      {
        chats()
      }
      <h2> game code {gameHash}</h2>
    </div>
  );
};

export default HangmanComponent;


