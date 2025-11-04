<script>
    let playlistURL = ""
    let artists = []

    const IDExtract = /(?:playlist\/|spotify:playlist:)([A-Za-z0-9]+)/

    async function fetchArtists(event) {
        event.preventDefault()
        artists = []

        const playlistID = IDExtract.exec(playlistURL)[1]
        console.log(playlistID)
        const response = await fetch(`/playlist/ruartists?id=${encodeURIComponent(playlistID)}`)
        if (!response.ok){
            throw new Error(`HTTP error status: ${response.status}`)
        }

        artists = await response.json()
    }
</script>

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
