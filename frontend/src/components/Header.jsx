
import React from "react";
import { Link } from "react-router-dom";

import { useLocation } from "react-router-dom";

export default function Header({ authenticated, onSignOut }) {
  const location = useLocation();
  if (!authenticated && location.pathname === "/") return null;
  return (
    <header>
      <nav>
        {authenticated && (
          <button className="signout" onClick={onSignOut}>Sign out</button>
        )}
      </nav>
    </header>
  );
}
