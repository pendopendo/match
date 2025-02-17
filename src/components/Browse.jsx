import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

function Browse() {
  const [profiles, setProfiles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  useEffect(() => {
    fetchProfiles();
  }, []);

  const fetchProfiles = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/recommendations', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to fetch profiles');
      }

      const data = await response.json();
      setProfiles(data.recommendations);
    } catch (err) {
      setError('Failed to load profiles');
    } finally {
      setLoading(false);
    }
  };

  const startChat = async (userId) => {
    try {
      const response = await fetch('http://localhost:8080/api/connections', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify({ user_id: userId }),
      });

      if (!response.ok) {
        throw new Error('Failed to start chat');
      }

      const data = await response.json();
      navigate(`/chat/${data.connection_id}`);
    } catch (err) {
      setError('Failed to start chat');
    }
  };

  if (loading) return <div className="browse-container">Loading...</div>;
  if (error) return <div className="browse-container error">{error}</div>;

  return (
    <div className="browse-container">
      <h2>Browse Profiles</h2>
      <div className="profiles-grid">
        {profiles.map(profile => (
          <div key={profile.user_id} className="profile-card">
            <img 
              src={profile.profile_picture || '/placeholder.png'} 
              alt={profile.name} 
              className="profile-card-image"
            />
            <div className="profile-card-info">
              <h3>{profile.name}</h3>
              <p>{profile.location}</p>
              <p className="profile-card-bio">{profile.bio}</p>
              <div className="profile-card-tags">
                {profile.interests?.map((interest, index) => (
                  <span key={index} className="tag">{interest}</span>
                ))}
              </div>
              <button 
                onClick={() => startChat(profile.user_id)}
                className="start-chat-button"
              >
                Start Chat
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export default Browse