import React, { useState } from 'react';
import './Chat.css'; // Import CSS file for styling (if using CSS-in-JS, skip this line)

export interface chatMessage {
  message: string,
  sender: string
}

interface ChatProps {
  chats: Array<chatMessage>
  sendMessage: (message: string) => void
  players: Array<string>
  playerIndex: number
  openChat: boolean
}

const ChatMessages: React.FC<ChatProps> = ({ chats, sendMessage, players, playerIndex, openChat }) => {
  const [messageValue, setInputValue] = useState('');
  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setInputValue(event.target.value);
  };
  const chatMessages = chats.map((element: chatMessage, _index: number) => {
    if (element.sender === players[playerIndex]) {
      let message: Array<any> = []
      for (let i = 27; i < element.message.length; i += 27) {
        message.push(element.message.substring(i - 27, i))
        message.push(
          <br />
        )
      }
      return (
        <div className="container">
          <p style={{ display: "inline-block" }}>{element.message}</p>
          <span className="time-right">{element.sender}</span>
        </div>
      )
    } else {
      return (
        <div className="container">
          <p>{element.message}</p>
          <span className="time-left">{element.sender}</span>
        </div>
      )
    }
  })


  if (openChat)
    return (
      <div className="chat-sidebar"> {/* Add a class to style the chat box container */}
        <h2>Chat Messages</h2>
        {
          chatMessages
        }
        <input
          value={messageValue}
          onKeyUp={(enter) => {
            if (enter.key === 'Enter') {
              sendMessage(messageValue)
              setInputValue("")
            }

          }} onChange={handleChange}></input>
        <button id="send-message" type="button" onClick={() => { sendMessage(messageValue); setInputValue("") }}></button>
      </div>
    );
  else return (<div />)
};

export default ChatMessages;
