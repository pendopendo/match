import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';

function ChatList() {
  const [connections, setConnections] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    fetchConnections();
  }, []);

  const fetchConnections = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/connections', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to fetch connections');
      }

      const data = await response.json();
      setConnections(data);
    } catch (err) {
      setError('Failed to load connections');
    } finally {
      setLoading(false);
    }
  };

  if (loading) return <div className="chat-list-container">Loading...</div>;
  if (error) return <div className="chat-list-container error">{error}</div>;

  return (
    <div className="chat-list-container">
      <h2>Your Chats</h2>
      {connections.length === 0 ? (
        <p>No chats yet. Start browsing to connect with people!</p>
      ) : (
        <div className="chat-list">
          {connections.map(connection => (
            <Link 
              key={connection.id} 
              to={`/chat/${connection.id}`}
              className="chat-list-item"
            >
              <div className="chat-list-item-avatar">
                <img 
                  src={connection.profile_picture || '/placeholder.png'} 
                  alt={connection.name} 
                />
              </div>
              <div className="chat-list-item-info">
                <h3>{connection.name}</h3>
                <p>{connection.last_message || 'No messages yet'}</p>
              </div>
              {connection.unread_count > 0 && (
                <div className="unread-badge">
                  {connection.unread_count}
                </div>
              )}
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}

export default ChatList;