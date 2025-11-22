import Header from "./components/Header";
import Dashboard from "./pages/Dashboard.tsx";
import Landing from "./pages/Landing";
import React from "react";
import { BrowserRouter as Router, Routes, Route, Navigate, useLocation, useNavigate, useNavigate as useNav } from "react-router-dom";

function RequireAuth({ children, authenticated }) {
  const location = useLocation();
  if (!authenticated) {
    return <Navigate to="/" state={{ from: location }} replace />;
  }
  return children;
}

// Custom fetch wrapper to catch 401s and redirect to landing
function useAuthFetch(setAuthenticated) {
  const navigate = useNav();
  React.useEffect(() => {
    const origFetch = window.fetch;
    window.fetch = async (...args) => {
      const res = await origFetch(...args);
      if (res.status === 401) {
        setAuthenticated(false);
        navigate("/", { replace: true });
      }
      return res;
    };
    return () => {
      window.fetch = origFetch;
    };
  }, [setAuthenticated, navigate]);
}

export default function App() {
  const [authenticated, setAuthenticated] = React.useState(false);
  const [checking, setChecking] = React.useState(true);

  React.useEffect(() => {
    (async () => {
      try {
        const res = await fetch("/api/me", { credentials: "include" });
        setAuthenticated(res.ok);
      } catch {
        setAuthenticated(false);
      } finally {
        setChecking(false);
      }
    })();
  }, []);

  const signOut = async () => {
    try {
      await fetch("/api/logout", { method: "POST", credentials: "include" });
    } catch (e) {}
    setAuthenticated(false);
  };

  if (checking) return null;

  return (
    <Router>
      <AuthFetchWrapper setAuthenticated={setAuthenticated} />
      <Header authenticated={authenticated} onSignOut={signOut} />
      <main>
        <Routes>
          <Route path="/dashboard" element={
            <RequireAuth authenticated={authenticated}>
              <Dashboard />
            </RequireAuth>
          } />
          <Route path="/" element={<Landing authenticated={authenticated} />} />
          <Route path="*" element={<Navigate to="/dashboard" replace />} />
        </Routes>
      </main>
    </Router>
  );
}

function AuthFetchWrapper({ setAuthenticated }) {
  useAuthFetch(setAuthenticated);
  return null;
}
