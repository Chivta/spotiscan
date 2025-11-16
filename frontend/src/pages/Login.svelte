<script>
    let { redirectLocation } = $props()
    let emailOrUsername = $state("")
    let password = $state("")
    let errorMessage = $state("")

    async function Login(event) {
        event.preventDefault()

        const response = await fetch("/api/login", {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            credentials: 'include',
            body: JSON.stringify({ emailOrUsername, password })
        })

        errorMessage = ""
        if (!response.ok) {
            let error = {}
            try {
                error = await response.json()
            } catch {}
            if (error.code === 'INVALID_CREDENTIALS') {
                errorMessage = 'Invalid credentials'
            } else {
                errorMessage = error.error || 'Login failed'
            }
            return
        }

        // success — backend sets HttpOnly cookie; redirect
        window.location.href = redirectLocation;
    }
</script>

<main>
    <form onsubmit={Login}>
        <h1>Log in</h1>
        {#if errorMessage}
            <div class="error">{errorMessage}</div>
        {/if}
        <label>
            Email or Username
            <input type="text" bind:value={emailOrUsername} required />
        </label>
        <label>
            Password
            <input type="password" bind:value={password} required />
        </label>
        <button type="submit">Log in</button>
    </form>
</main>

<style>

    .error {
        color: red;
        text-align: center;
    }

    main {
        display: flex;
        justify-content: center;
        align-items: center;
        min-height: 80vh;
    }

    form {
        display: grid;
        gap: 1rem;
        width: 100%;
        max-width: 400px;
        padding: 2rem;
        border: 1px solid #ddd;
        border-radius: 8px;
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    }

    h1 {
        margin: 0 0 1rem;
        text-align: center;
        font-size: 1.5rem;
    }

    label {
        display: grid;
        font-size: 0.9rem;
        font-weight: 500;
    }

    input {
        padding: 0.5rem;
        border: 1px solid #ccc;
        border-radius: 4px;
        font-size: 1rem;
    }

    input:focus {
        outline: none;
        border-color: #007bff;
    }

    button {
        padding: 0.75rem;
        background: #007bff;
        color: white;
        border: none;
        border-radius: 4px;
        font-size: 1rem;
        cursor: pointer;
        transition: background 0.2s;
    }

    button:hover {
        background: #00d92f;
    }
</style>