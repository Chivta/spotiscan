import { useState, useEffect } from "react";
import { BrowserRouter as Router, Routes, Route, Navigate, useNavigate } from "react-router-dom";
import Landing from "./pages/Landing.tsx";
import Dashboard from "./pages/Dashboard.tsx";
import AuthPage from "./pages/AuthPage.tsx";
import Header from "./components/Header.tsx";
import type { User } from "./types/models.ts";

function RootRoute({ onUser }: { onUser: (user: User) => void }) {
  const [checked, setChecked] = useState(false);
  const [authed, setAuthed] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    fetch("/api/me", { credentials: "include" })
      .then(res => {
        if (res.ok) return res.json();
        throw new Error("unauthenticated");
      })
      .then((user: User) => {
        if (user.userRole !== "anon") {
          onUser(user);
          setAuthed(true);
          navigate("/dashboard", { replace: true });
        } else {
          setChecked(true);
        }
      })
      .catch(() => setChecked(true));
  }, []);

  if (authed) return null;
  if (!checked) return null;
  return <Landing />;
}

function AppRoutes() {
  const [user, setUser] = useState<User | null>(null);
  const navigate = useNavigate();

  const handleLogout = async () => {
    await fetch("/api/auth/logout", { method: "POST", credentials: "include" });
    setUser(null);
    navigate("/", { replace: true });
  };

  return (
    <Routes>
      <Route path="/" element={<RootRoute onUser={setUser} />} />
      <Route path="/signup" element={<AuthPage initialMode="signup" />} />
      <Route path="/login" element={<AuthPage initialMode="login" />} />
      <Route
        path="/dashboard"
        element={
          user && user.userRole !== "anon" ? (
            <>
              <Header user={user} onLogout={handleLogout} />
              <Dashboard />
            </>
          ) : (
            <Navigate to="/" replace />
          )
        }
      />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}

export default function App() {
  return (
    <Router>
      <main>
        <AppRoutes />
      </main>
    </Router>
  );
}
