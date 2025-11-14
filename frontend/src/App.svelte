<script>
  import Router from 'svelte-spa-router';
  import Scanner from './pages/Scanner.svelte';
  import Singup from './pages/Signup.svelte';

  const routes = {
    '/scanner': Scanner,
    '/signup': Singup,
  };
</script>

<header>
  <nav>
    <a href="#/">Home</a>
    <a href="#/scanner">Scanner</a>
    <a href="#/signup">Sing up</a>
  </nav>
</header>

<main>
  <Router {routes}>
    <div slot="notFound">
      <h1>Welcome to Spotiscan</h1>
      <p>Your home page content here</p>
    </div>
  </Router>
</main>