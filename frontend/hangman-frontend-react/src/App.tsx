import React, { useEffect, useState } from 'react';
import './App.css';
import HangmanComponent from './hangman';
import axios, { Axios } from 'axios'
import Cookies from 'js-cookie';

export const TabTitle = (newTitle: string) => {
  return (document.title = newTitle);
};

export const getHashCookie = () => {
  let hash = Cookies.get('hash')
  if (hash === undefined || hash === '' || hash === 'undefined') {
    return ""
  } else {
    return hash
  }
}


export const setHashCookie = (hash: string) => {
  Cookies.set('hash', hash)
}




function App() {
  // const _url = "https://hangman-backend-geoffrey.com"
  // const _url = "http://18.189.248.181:8080"
  const _url = "http://localhost:8080"

  const [gameChoice, setGameChoice] = useState<number>(-1)
  const [games, setGames] = useState<number>(-1)
  const [reconnect, setReconnect] = useState<boolean>(false)
  TabTitle("Geoffrey's Hangman Server")
  let hash = getHashCookie()

  const exitGame = () => {
    setGameChoice(-1)
    setReconnect(false)
    axios.get(_url + '/get_games',).then((response) => {
      console.log("response\n" + response)
      setGames(response.data)
    })
  }


  useEffect(
    () => {
      axios.get(_url + '/get_games',).then((response) => {
        console.log("response\n" + response)
        setGames(response.data)
      })
      if (hash !== '') {
        axios.get(_url + '/valid/' + hash,).then((response) => {
          console.log("response\n" + response.data)
          if (response.data === -1) {

            setReconnect(false)
            setGameChoice(-1)
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
      response.json().then((obj: { length: number }) => {
        setGames(obj.length)
        setGameChoice(games)
      })
    })
  }
  const selectGame = (index: number, reconnect: boolean) => {
    return (
      <HangmanComponent gameIndex={index} reconnect={reconnect} hash={hash ? hash : ""} reset={exitGame}></HangmanComponent>
    );
  }



  const needToSelectGame = () => {
    let ary = []
    for (let i = 0; i < games; i++) {
      ary.push((<div key={i}><button id={"join-game-" + i} key={i} onClick={() => { setGameChoice(i) }}>Join Game</button></div>))
    }
    return ary
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
        gameChoice === -1 ?
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
