import React from "react";
import { useLocation } from "react-router-dom";

interface HeaderProps {
  authenticated: boolean;
  onSignOut: () => void;
}

export default function Header({ authenticated, onSignOut }: HeaderProps) {
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
