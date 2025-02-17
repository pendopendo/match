import { useState, useEffect, useRef } from 'react';
import { useParams } from 'react-router-dom';

function Chat() {
  const { connectionId } = useParams();
  const [messages, setMessages] = useState([]);
  const [newMessage, setNewMessage] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const [ws, setWs] = useState(null);
  const messagesEndRef = useRef(null);

  useEffect(() => {
    // Fetch existing messages
    fetchMessages();
    
    // Setup WebSocket connection
    const websocket = new WebSocket(`ws://localhost:8080/ws/chat/${connectionId}`);
    
    websocket.onopen = () => {
      console.log('Connected to chat');
    };

    websocket.onmessage = (event) => {
      const message = JSON.parse(event.data);
      setMessages(prev => [...prev, message]);
    };

    websocket.onerror = (error) => {
      console.error('WebSocket error:', error);
      setError('Connection error');
    };

    websocket.onclose = () => {
      console.log('Disconnected from chat');
    };

    setWs(websocket);

    return () => {
      websocket.close();
    };
  }, [connectionId]);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const fetchMessages = async () => {
    try {
      const response = await fetch(`http://localhost:8080/api/connections/${connectionId}/messages`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to fetch messages');
      }

      const data = await response.json();
      setMessages(data);
    } catch (err) {
      setError('Failed to load messages');
    } finally {
      setLoading(false);
    }
  };

  const sendMessage = (e) => {
    e.preventDefault();
    if (!newMessage.trim()) return;

    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({
        connectionId: parseInt(connectionId),
        content: newMessage.trim()
      }));
      setNewMessage('');
    } else {
      setError('Connection lost. Please refresh the page.');
    }
  };

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  if (loading) return <div className="chat-container">Loading...</div>;
  if (error) return <div className="chat-container error">{error}</div>;

  return (
    <div className="chat-container">
      <div className="messages-container">
        {messages.map((message, index) => (
          <div
            key={index}
            className={`message ${message.sender_id === parseInt(localStorage.getItem('userId')) ? 'sent' : 'received'}`}
          >
            <div className="message-content">{message.content}</div>
            <div className="message-time">
              {new Date(message.created_at).toLocaleTimeString()}
            </div>
          </div>
        ))}
        <div ref={messagesEndRef} />
      </div>
      <form onSubmit={sendMessage} className="message-form">
        <input
          type="text"
          value={newMessage}
          onChange={(e) => setNewMessage(e.target.value)}
          placeholder="Type a message..."
          className="message-input"
        />
        <button type="submit" className="send-button">Send</button>
      </form>
    </div>
  );
}

export default Chat;