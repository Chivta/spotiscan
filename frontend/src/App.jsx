import Header from "./components/Header";
import Scanner from "./pages/Scanner";
import Auth from "./pages/Auth";
import React from "react";
import { BrowserRouter as Router, Routes, Route, Navigate, useLocation, useNavigate } from "react-router-dom";

function RequireAuth({ children, authenticated }) {
  const location = useLocation();
  if (!authenticated) {
    return <Navigate to="/auth" state={{ from: location }} replace />;
  }
  return children;
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
      <Header authenticated={authenticated} onSignOut={signOut} />
      <main>
        <Routes>
          <Route path="/scanner" element={
            <RequireAuth authenticated={authenticated}>
              <Scanner />
            </RequireAuth>
          } />
          <Route path="/auth" element={<Auth />} />
          <Route path="*" element={<Navigate to="/scanner" replace />} />
        </Routes>
      </main>
    </Router>
  );
}
