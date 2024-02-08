import React, { useEffect, useState } from 'react';
import './App.css';
import HangmanComponent from './hangman';
import axios from 'axios'





function App() {
  const url = "http://localhost:8000"

  const [gameChoice, setGameChoice] = useState<number>(-1)
  const [games, setGames] = useState<number>(-1)

  useEffect(
    () => {
      axios.get(url + '/get_games').then((response) => {
        console.log("response\n" + response)
        setGames(response.data)
      })

    }, []
  )
  const sendNewGame = () => {
    fetch(url + "/new_game").then((response: any) => {
      response.json().then((obj: { length: number }) => {
        setGames(obj.length)
      })
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
      ary.push((<button onClick={() => { setGameChoice(i) }}>Join Game</button>))
    }
    return ary
  }


  return (

    <div className="App">
      {
        gameChoice === -1 ?
          (
            <div>{needToSelectGame()}
              <button onClick={() => {
                sendNewGame()
                setGameChoice(games)
              }}>New Game</button>
            </div>
          )
          : selectGame(gameChoice)

      }
    </div>
  );
}

export default App;
