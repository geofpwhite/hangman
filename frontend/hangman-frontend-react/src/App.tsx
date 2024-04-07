import React, { useEffect, useState } from 'react';
import './App.css';
import HangmanComponent from './hangman';
import axios, { Axios } from 'axios'
import Cookies from 'js-cookie';

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
  const _url = "https://hangman-backend-geoffrey.com"
  // const _url = "http://18.189.248.181:8080"
  // const _url = "http://localhost:8080"

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
    }, []
  )
  const sendNewGame = () => {
    fetch(_url + "/new_game").then((response: any) => {
      response.json().then((obj: { gameHash: string }) => {
        setGameChoice(obj.gameHash)
      })
    })
  }
  const selectGame = (gameHash: string, reconnect: boolean) => {
    return (
      <HangmanComponent gameHash={gameHash} reconnect={reconnect} hash={hash ? hash : ""} reset={exitGame}></HangmanComponent>
    );
  }
  const typeInGameCode = () => {
    return (
      <div>
        <div className="game-code">
          <div>
            <label htmlFor="gameCode">
              We need a new word </label>
            <input
              type="text"
              id="gameCode"
              value={gameChoice}
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
    let input = typeInGameCode()
    return input
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
                <button onClick={() => {
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
