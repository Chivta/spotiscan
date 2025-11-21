import { useEffect } from "react";

export default function Auth() {
  useEffect(() => {
    window.location.href = "/api/auth";
  }, []);
  return (
    <main>
      <h1>Redirecting to Spotify...</h1>
    </main>
  );
}