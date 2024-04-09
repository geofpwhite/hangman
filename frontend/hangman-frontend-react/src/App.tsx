import React, { useEffect, useState } from 'react';
import './App.css';
import HangmanComponent from './hangman';
import axios, { Axios } from 'axios'
import Cookies from 'js-cookie';
import { Route } from 'react-router-dom';

export const TabTitle = (newTitle: string) => {
  return (document.title = newTitle);
};

export const getHashCookies = () => {
  let hash = Cookies.get('hash')
  let gameHash = Cookies.get('gameHash')
  return [hash, gameHash]
}


export const setHashCookie = (hash: string) => {
  Cookies.set('hash', hash)
}
export const setGameHashCookie = (hash: string) => {
  Cookies.set('gameHash', hash)
}




function App() {
  const _url = "http://localhost:8080"

  const [gameChoice, setGameChoice] = useState<string>("")
  const [gameCodeInput, setGameCodeInput] = useState<string>("")
  const [reconnect, setReconnect] = useState<boolean>(false)
  TabTitle("Geoffrey's Hangman Server")
  let hashes = getHashCookies()
  let hash = hashes[0]
  let gameHash = hashes[1]

  const exitGame = () => {
    setGameChoice("")
    setReconnect(false)
    axios.get(_url + '/get_games',).then((response) => {
      console.log("response\n" + response)
    })
  }


  useEffect(
    () => {
      axios.get(_url + '/get_games',).then((response) => {
        console.log("response\n" + response)
      })
      if (hash !== '') {
        axios.get(_url + '/valid/' + hash,).then((response) => {
          console.log("response\n" + response.data)
          if (response.data === -1) {

            setReconnect(false)
            setGameChoice('')
          } else {
            setReconnect(true)
            setGameChoice(response.data)
          }
        })
      }
    }, []
  )
  const sendNewGame = () => {
    fetch(_url + "/new_game").then((response: any) => {
      response.json().then((obj: { gameHash: string }) => {
        setGameChoice(obj.gameHash)
      })
    })
  }
  const selectGame = (gameHash: string, reconnect: boolean) => (
    // <Route path="/:gameHash" element={<HangmanComponent gameHash={gameHash} reconnect={reconnect} hash={hash ? hash : ""} reset={exitGame}></HangmanComponent>} ></Route>
    <HangmanComponent gameHash={gameHash} reconnect={reconnect} hash={hash ? hash : ""} reset={exitGame}></HangmanComponent>
  )
  const typeInGameCode = () => {
    return (
      <div>
        <div className="game-code">
          <div>
            <label htmlFor="gameCode">
              type in a game code to join</label>
            <input
              type="text"
              id="gameCode"
              value={gameCodeInput}
              onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
                setGameCodeInput(event.target.value);
              }}
              placeholder="Type here..."
            />
            <button type="button" onClick={() => setGameChoice(gameCodeInput)}>Submit</button>
          </div>
        </div>
      </div>
    )
  }



  const needToSelectGame = () => {
    return typeInGameCode()
  }

  if (reconnect && hash !== "") {
    return (
      <div className="App">
        {
          selectGame(gameChoice, true)
        }
      </div>
    );
  }

  return (
    <div className="App">
      {
        gameChoice === "" ?
          (
            <div style={{ padding: '30%' }}>
              <div>
                {needToSelectGame()}
              </div>
              <div>
                <button id="new-game" key="new-game" onClick={() => {
                  sendNewGame()
                }}>New Game</button>
              </div>
            </div>
          )
          : selectGame(gameChoice, false)
      }
    </div>
  );
}

export default App;
