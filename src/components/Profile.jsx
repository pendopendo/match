import { useState, useEffect } from 'react';

function Profile() {
  const [profile, setProfile] = useState({
    name: '',
    bio: '',
    location: '',
    profile_picture: ''
  });
  
  const [bio, setBio] = useState({
    interests: [],
    hobbies: [],
    music_preferences: [],
    food_preferences: [],
    looking_for: []
  });
  const [recommendations, setRecommendations] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [editMode, setEditMode] = useState(false);

  // Form states
  const [formData, setFormData] = useState({
    name: '',
    bio: '',
    profilePicture: '',
    location: '',
    interests: [],
    hobbies: [],
    musicPreferences: [],
    foodPreferences: [],
    lookingFor: []
  });

  useEffect(() => {
    const fetchData = async () => {
        try {
            const token = localStorage.getItem('token');
            if (!token) {
                navigate('/login');
                return;
            }

            const [profileRes, bioRes] = await Promise.all([
                fetch('http://localhost:8080/api/me/profile', {
                    headers: {
                        'Authorization': `Bearer ${token}`
                    }
                }),
                fetch('http://localhost:8080/api/me/bio', {
                    headers: {
                        'Authorization': `Bearer ${token}`
                    }
                })
            ]);

            if (!profileRes.ok || !bioRes.ok) {
                console.error('Profile Response:', await profileRes.text());
                console.error('Bio Response:', await bioRes.text());
                throw new Error('Failed to fetch data');
            }

            const profileData = await profileRes.json();
            const bioData = await bioRes.json();

            setProfile(profileData);
            setBio(bioData);
            setLoading(false);
        } catch (error) {
            console.error('Error fetching data:', error);
            setLoading(false);
            setError(error.message);
        }
    };

    fetchData();
}, [navigate]);

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const handleArrayInputChange = (e, field) => {
    const value = e.target.value.split(',').map(item => item.trim());
    setFormData(prev => ({
      ...prev,
      [field]: value
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    try {
      // Update profile
      const profileRes = await fetch('http://localhost:8080/api/me/profile', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify({
          name: formData.name,
          bio: formData.bio,
          profile_picture: formData.profilePicture,
          location: formData.location
        }),
      });

      // Update bio
      const bioRes = await fetch('http://localhost:8080/api/me/bio', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify({
          interests: formData.interests,
          hobbies: formData.hobbies,
          music_preferences: formData.musicPreferences,
          food_preferences: formData.foodPreferences,
          looking_for: formData.lookingFor
        }),
      });

      if (profileRes.ok && bioRes.ok) {
        setEditMode(false);
        fetchProfileData();
      } else {
        throw new Error('Failed to update profile');
      }
    } catch (err) {
      setError('Failed to update profile');
    }
  };

  if (loading) return <div className="profile-container">Loading...</div>;
  if (error) return <div className="profile-container error">{error}</div>;

  return (
    <div className="profile-container">
      <h2>My Profile</h2>
      
      {!editMode ? (
        <div className="profile-view">
          <div className="profile-info">
            <img
              src={profile?.profile_picture || '/placeholder.png'}
              alt="Profile"
              className="profile-picture"
            />
            <div className="profile-details">
              <h3>{profile?.name || 'No name set'}</h3>
              <p>{profile?.location || 'No location set'}</p>
              <p>{profile?.bio || 'No bio set'}</p>
            </div>
          </div>

          <div className="bio-info">
            <h3>About Me</h3>
            <div className="bio-section">
              <h4>Interests</h4>
              <p>{bio?.interests?.join(', ') || 'None set'}</p>
            </div>
            <div className="bio-section">
              <h4>Hobbies</h4>
              <p>{bio?.hobbies?.join(', ') || 'None set'}</p>
            </div>
            <div className="bio-section">
              <h4>Music Preferences</h4>
              <p>{bio?.music_preferences?.join(', ') || 'None set'}</p>
            </div>
            <div className="bio-section">
              <h4>Food Preferences</h4>
              <p>{bio?.food_preferences?.join(', ') || 'None set'}</p>
            </div>
            <div className="bio-section">
              <h4>Looking For</h4>
              <p>{bio?.looking_for?.join(', ') || 'None set'}</p>
            </div>
          </div>

          <button onClick={() => setEditMode(true)} className="edit-button">
            Edit Profile
          </button>

          {recommendations.length > 0 && (
            <div className="recommendations">
              <h3>Recommended Matches</h3>
              <div className="recommendations-list">
                {recommendations.map(id => (
                  <div key={id} className="recommendation-card">
                    User {id}
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      ) : (
        <form onSubmit={handleSubmit} className="profile-form">
          <div className="form-group">
            <label>Name:</label>
            <input
              type="text"
              name="name"
              value={formData.name}
              onChange={handleInputChange}
            />
          </div>

          <div className="form-group">
            <label>Bio:</label>
            <textarea
              name="bio"
              value={formData.bio}
              onChange={handleInputChange}
            />
          </div>

          <div className="form-group">
            <label>Profile Picture URL:</label>
            <input
              type="text"
              name="profilePicture"
              value={formData.profilePicture}
              onChange={handleInputChange}
            />
          </div>

          <div className="form-group">
            <label>Location:</label>
            <input
              type="text"
              name="location"
              value={formData.location}
              onChange={handleInputChange}
            />
          </div>

          <div className="form-group">
            <label>Interests (comma-separated):</label>
            <input
              type="text"
              value={formData.interests.join(', ')}
              onChange={(e) => handleArrayInputChange(e, 'interests')}
            />
          </div>

          <div className="form-group">
            <label>Hobbies (comma-separated):</label>
            <input
              type="text"
              value={formData.hobbies.join(', ')}
              onChange={(e) => handleArrayInputChange(e, 'hobbies')}
            />
          </div>

          <div className="form-group">
            <label>Music Preferences (comma-separated):</label>
            <input
              type="text"
              value={formData.musicPreferences.join(', ')}
              onChange={(e) => handleArrayInputChange(e, 'musicPreferences')}
            />
          </div>

          <div className="form-group">
            <label>Food Preferences (comma-separated):</label>
            <input
              type="text"
              value={formData.foodPreferences.join(', ')}
              onChange={(e) => handleArrayInputChange(e, 'foodPreferences')}
            />
          </div>

          <div className="form-group">
            <label>Looking For (comma-separated):</label>
            <input
              type="text"
              value={formData.lookingFor.join(', ')}
              onChange={(e) => handleArrayInputChange(e, 'lookingFor')}
            />
          </div>

          <div className="form-buttons">
            <button type="submit">Save</button>
            <button type="button" onClick={() => setEditMode(false)} className="cancel-button">
              Cancel
            </button>
          </div>
        </form>
      )}
    </div>
  );
}

export default Profile;