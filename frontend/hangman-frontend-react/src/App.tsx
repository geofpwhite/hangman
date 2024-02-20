import React, { useEffect, useState } from 'react';
import './App.css';
import HangmanComponent from './hangman';
import axios, { Axios } from 'axios'
export const TabTitle = (newTitle: string) => {
  return (document.title = newTitle);
};




function App() {
  const _url = "https://hangman-backend-geoffrey.com"
  // const _url = "http://18.189.248.181:8080"

  const [gameChoice, setGameChoice] = useState<number>(-1)
  const [games, setGames] = useState<number>(-1)
  TabTitle("Geoffrey's Hangman Server")

  useEffect(
    () => {

      axios.get(_url + '/get_games',).then((response) => {
        console.log("response\n" + response)
        setGames(response.data)
      })

    }, []
  )
  const sendNewGame = () => {
    fetch(_url + "/new_game").then((response: any) => {
      response.json().then((obj: { length: number }) => {
        setGames(obj.length)
      })
    }).then(() => {
      setGameChoice(games)
    })
  }
  const selectGame = (index: number) => {
    return (
      <HangmanComponent gameIndex={index}></HangmanComponent>
    );
  }



  const needToSelectGame = () => {
    let ary = []
    for (let i = 0; i < games; i++) {
      ary.push((<div><button onClick={() => { setGameChoice(i) }}>Join Game</button></div>))
    }
    return ary
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
                <button onClick={() => {
                  sendNewGame()
                }}>New Game</button>
              </div>
            </div>
          )
          : selectGame(gameChoice)
      }
    </div>
  );
}

export default App;
