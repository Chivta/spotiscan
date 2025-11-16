<script>
  export let authenticated = false;
  export let onSignOut = () => {};
</script>

<header>
  <nav>
    <a href="#/">Home</a>
    <a href="#/scanner">Scanner</a>

    {#if !authenticated}
      <a href="#/signup">Sign up</a>
      <a href="#/login">Login</a>
    {:else}
      <button class="signout" on:click={onSignOut}>Sign out</button>
    {/if}
  </nav>
</header>

<style>
  header {
    background-color: #222;
    padding: 1rem;
  }

  nav {
    display: flex;
    gap: 1rem;
    justify-content: center;
  }

  a {
    color: white;
    text-decoration: none;
    font-weight: bold;
  }

  a:hover {
    text-decoration: underline;
  }

  .signout {
    background: transparent;
    color: white;
    border: 1px solid rgba(255,255,255,0.15);
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    cursor: pointer;
  }
  .signout:hover {
    background: rgba(255,255,255,0.05);
  }
</style>