
import React from "react";
import { Link } from "react-router-dom";

export default function Header({ authenticated, onSignOut }) {
  return (
    <header>
      <nav>
        <Link to="/">Home</Link>
        <Link to="/scanner">Scanner</Link>
        {!authenticated ? (
          <Link to="/auth">Login</Link>
        ) : (
          <button className="signout" onClick={onSignOut}>Sign out</button>
        )}
      </nav>
    </header>
  );
}
