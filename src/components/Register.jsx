import { useState } from 'react';
import { useNavigate } from 'react-router-dom';

function Register() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    if (password.length < 6) {
      setError('Password must be at least 6 characters long');
      return;
    }

    try {
      const response = await fetch('http://localhost:8080/api/register', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json'
        },
        body: JSON.stringify({ 
          email: email.trim(),
          password: password
        }),
      });

      if (!response.ok) {
        const errorData = await response.text();
        throw new Error(errorData || 'Registration failed');
      }

      // Registration successful
      setError('');
      navigate('/login');
    } catch (err) {
      console.error('Registration error:', err);
      if (err instanceof TypeError && err.message.includes('Failed to fetch')) {
        setError('Unable to connect to the server. Please make sure the backend server is running.');
      } else if (err.message.includes('Email already exists')) {
        setError('This email is already registered. Please use a different email.');
      } else {
        setError('Registration failed. Please try again.');
      }
    }
  };

  return (
    <div className="auth-container">
      <h2>Register</h2>
      {error && <div className="error">{error}</div>}
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label>Email:</label>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            placeholder="Enter your email"
            className={error && error.includes('email') ? 'error' : ''}
          />
        </div>
        <div className="form-group">
          <label>Password:</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            placeholder="Enter your password (min. 6 characters)"
            minLength="6"
            className={error && error.includes('password') ? 'error' : ''}
          />
        </div>
        <button type="submit">Register</button>
      </form>
    </div>
  );
}

export default Register;