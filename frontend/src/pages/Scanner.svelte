<script lang="ts">
    let playlistURL: string = ""
    let artists: string[] = []

    const IDExtract: RegExp = /(?:playlist\/|spotify:playlist:)([A-Za-z0-9]+)/

    async function fetchArtists(event: Event) {
        event.preventDefault()
        artists = []
        
        // Check if URL matches the expected pattern
        if (!IDExtract.test(playlistURL)) {
            alert("Invalid Spotify playlist URL format")
            return
        }

        const match = IDExtract.exec(playlistURL)
        const playlistID = match![1] 
        console.log(playlistID)
        const response = await fetch(`/api/playlist/ruartists?id=${encodeURIComponent(playlistID)}`)
        if (!response.ok){
            throw new Error(`HTTP error status: ${response.status}`)
        }

        artists = await response.json()
    }
</script>

<main>
    <h1>Paste spotify playlist url</h1>
    <form>
        <input bind:value={playlistURL} type="text">
        <button on:click={fetchArtists} placeholder="Playlist url">Get ru artists</button>
    </form>
    <ul>
        {#each artists as artist}
        <li>{artist}</li>
        {/each}
    </ul>
</main>