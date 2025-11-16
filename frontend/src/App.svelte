<script>
  import Router from 'svelte-spa-router';
  import Scanner from './pages/Scanner.svelte';
  import Signup from './pages/Signup.svelte';
  import Login from './pages/Login.svelte';
  import Header from './components/Header.svelte';
  import { onMount } from 'svelte';

  const routes = {
    '/scanner': Scanner,
    '/signup': Signup,
    '/login': Login,
  };

  let authenticated = false;
  let checking = true;

  onMount(async () => {
    try {
      const res = await fetch('/api/me', { credentials: 'include' }); // endpoint returns 200 if authenticated
      authenticated = res.ok;
    } catch {
      authenticated = false;
    } finally {
      checking = false;
    }
  });

  async function signOut() {
    try {
      await fetch('/api/logout', { method: 'POST', credentials: 'include' });
    } catch (e) {
      // ignore network errors, still flip UI
    }
    authenticated = false;
  }
</script>

<Header {authenticated} onSignOut={signOut} />

<main>
  <Router {routes}/>
</main>