<script lang="ts">
    let username: string = ""
    let email: string = ""
    let password: string = ""

    async function signUp(event: Event) {
        event.preventDefault()
        
        try {
            const response = await fetch('/api/signup', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ username, email, password })
            })

            if (!response.ok) {
                const error = await response.json()
                alert(error.message || 'Signup failed')
                return
            }

            const data = await response.json()
            console.log('Signup successful:', data)
            

            // TODO: Redirect and save cookie
            
        } catch (error) {
            console.error('Signup error:', error)
            alert('Something went wrong')
        }
    }
</script>

<main>
    <form on:submit={signUp}>
        <h1>Sign Up</h1>
        <label>
            Username
            <input type="text" bind:value={username} required />
        </label>
        <label>
            Email
            <input type="email" bind:value={email} required />
        </label>
        <label>
            Password
            <input type="password" bind:value={password} required />
        </label>
        <button type="submit">Sign up</button>
    </form>
</main>

<style>
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
        gap: 0.25rem;
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