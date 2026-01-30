package spotify_client

func NewSpotifyClient(spotifyId, spotifySecret string) *SpotifyClient {
	return &SpotifyClient{
		spotifyId:     spotifyId,
		spotifySecret: spotifySecret,
	}
}

type SpotifyClient struct {
	spotifyId     string
	spotifySecret string
}