import { useState, useEffect } from "react";
import { BrowserRouter as Router, Routes, Route, Navigate, useNavigate } from "react-router-dom";
import Landing from "./pages/Landing.tsx";
import Dashboard from "./pages/Dashboard.tsx";
import AdminPage from "./pages/AdminPage.tsx";
import AuthPage from "./pages/AuthPage.tsx";
import Header from "./components/Header.tsx";
import type { User } from "./types/models.ts";
import { LanguageProvider } from "./context/LanguageContext.tsx";

function RootRoute() {
  const [checked, setChecked] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    fetch("/api/me", { credentials: "include" })
      .then(res => {
        if (res.ok) return res.json();
        throw new Error("unauthenticated");
      })
      .then((user: User) => {
        if (user.userRole !== "anon") {
          navigate("/dashboard", { replace: true });
        } else {
          setChecked(true);
        }
      })
      .catch(() => setChecked(true));
  }, []);

  if (!checked) return null;
  return <Landing />;
}

function ProtectedRoute({ requiredRole, render }: {
  requiredRole?: "admin";
  render: (user: User, onLogout: () => void) => React.ReactNode;
}) {
  const [user, setUser] = useState<User | null>(null);
  const [checked, setChecked] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    fetch("/api/me", { credentials: "include" })
      .then(res => res.ok ? res.json() : Promise.reject())
      .then((u: User) => {
        if (u.userRole !== "anon" && (!requiredRole || u.userRole === requiredRole)) {
          setUser(u);
        }
        setChecked(true);
      })
      .catch(() => setChecked(true));
  }, []);

  const handleLogout = async () => {
    await fetch("/api/auth/logout", { method: "POST", credentials: "include" });
    navigate("/", { replace: true });
  };

  if (!checked) return null;
  if (!user) return <Navigate to="/" replace />;
  return <>{render(user, handleLogout)}</>;
}

function AppRoutes() {
  return (
    <Routes>
      <Route path="/" element={<RootRoute />} />
      <Route path="/signup" element={<AuthPage initialMode="signup" />} />
      <Route path="/login" element={<AuthPage initialMode="login" />} />
      <Route
        path="/dashboard"
        element={
          <ProtectedRoute render={(user, onLogout) => (
            <>
              <Header user={user} onLogout={onLogout} />
              <Dashboard />
            </>
          )} />
        }
      />
      <Route
        path="/admin"
        element={
          <ProtectedRoute requiredRole="admin" render={(user, onLogout) => (
            <>
              <Header user={user} onLogout={onLogout} />
              <AdminPage />
            </>
          )} />
        }
      />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}

export default function App() {
  return (
    <LanguageProvider>
      <Router>
        <main>
          <AppRoutes />
        </main>
      </Router>
    </LanguageProvider>
  );
}
