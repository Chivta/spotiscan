import React from "react";
import { useLocation, useNavigate } from "react-router-dom";

interface HeaderProps {
  authenticated: boolean;
  onSignOut: () => void;
}

export default function Header({ authenticated, onSignOut }: HeaderProps) {
  const location = useLocation();
  const navigate = useNavigate();

  const handleSignOut = async () => {
    await onSignOut();
    navigate("/", { replace: true });
  };

  // Only show logout button on dashboard
  if (!authenticated || location.pathname !== "/dashboard") return null;

  return (
    <button
      onClick={handleSignOut}
      style={{
        position: "fixed",
        top: 12,
        right: 16,
        zIndex: 10,
        padding: "8px 16px",
        background: "#caffdcff",
        color: "#000",
        border: "none",
        borderRadius: 50,
        fontWeight: 600,
        fontSize: 12,
        cursor: "pointer",
        transition: "all 0.2s ease",
        boxShadow: "0 2px 8px 0 rgba(0,0,0,0.08)",
      }}
    >
      Sign Out
    </button>
  );
}
