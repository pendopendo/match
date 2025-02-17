import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { useState, useEffect } from 'react';
import Login from './components/Login';
import Register from './components/Register';
import Profile from './components/Profile';
import Browse from './components/Browse';
import Chat from './components/Chat';
import ChatList from './components/ChatList';
import Navbar from './components/Navbar';
import './App.css';

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [user, setUser] = useState(null);

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (token) {
      setIsAuthenticated(true);
      // TODO: Fetch user data
    }
  }, []);

  const handleLogout = () => {
    localStorage.removeItem('token');
    setIsAuthenticated(false);
    setUser(null);
  };

  return (
    <Router>
      <div className="app">
        <Navbar isAuthenticated={isAuthenticated} onLogout={handleLogout} />
        <Routes>
          <Route path="/login" element={!isAuthenticated ? <Login setIsAuthenticated={setIsAuthenticated} /> : <Navigate to="/profile" />} />
          <Route path="/register" element={!isAuthenticated ? <Register /> : <Navigate to="/profile" />} />
          <Route path="/profile" element={isAuthenticated ? <Profile user={user} /> : <Navigate to="/login" />} />
          <Route path="/browse" element={isAuthenticated ? <Browse /> : <Navigate to="/login" />} />
          <Route path="/chats" element={isAuthenticated ? <ChatList /> : <Navigate to="/login" />} />
          <Route path="/chat/:connectionId" element={isAuthenticated ? <Chat /> : <Navigate to="/login" />} />
          <Route path="/" element={<Navigate to={isAuthenticated ? "/profile" : "/login"} />} />
        </Routes>
      </div>
    </Router>
  );
}

export default App;