import { Link } from 'react-router-dom';

function Navbar({ isAuthenticated, onLogout }) {
  return (
    <nav className="navbar">
      <div className="nav-brand">Match Me</div>
      <div className="nav-links">
        {isAuthenticated ? (
          <>
            <Link to="/profile">Profile</Link>
            <button onClick={onLogout} className="nav-button">Logout</button>
          </>
        ) : (
          <>
            <Link to="/login">Login</Link>
            <Link to="/register">Register</Link>
          </>
        )}
      </div>
    </nav>
  );
}

export default Navbar;