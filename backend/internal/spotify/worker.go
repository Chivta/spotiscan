package spotify

import (
	"context"

	"github.com/chivta/ruscan/internal/shared/models"
	"github.com/chivta/ruscan/internal/shared/queue"
	"github.com/rs/zerolog/log"
)

func NewSpotifyGatewayWorker(queue *queue.Client, spotifyClient *SpotifyClient) *SpotifyGatewayWorker {
	return &SpotifyGatewayWorker{queue: queue, spotifyClient: spotifyClient}
}

type SpotifyGatewayWorker struct {
	queue         *queue.Client
	spotifyClient *SpotifyClient
}

func (w *SpotifyGatewayWorker) Start(ctx context.Context) error {
	err := w.queue.DeclareQueue(models.SpotifyQueueName)
	if err != nil {
		return err
	}

	deliveries, err := w.queue.ConsumeScanJobs(ctx, models.SpotifyQueueName)
	if err != nil {
		return err
	}

	for delivery := range deliveries {
		job := delivery.Job
		err := w.processJob(ctx, job)
		if err != nil {
			delivery.Msg.Nack(false, true) // requeue on error
		} else {
			delivery.Msg.Ack(false)
		}
	}
	return nil
}

func (w *SpotifyGatewayWorker) processJob(ctx context.Context, job *queue.ScanJob) error {
	var fetchFunc func(context.Context, string) (*models.Content, error)
	switch job.ResourceType {
	case queue.ResourceType_PLAYLIST_ID:
		fetchFunc = w.spotifyClient.GetSpotifyPlaylist
	case queue.ResourceType_TRACK_ID:
		fetchFunc = w.spotifyClient.GetSpotifyTrack
	case queue.ResourceType_ALBUM_ID:
		fetchFunc = w.spotifyClient.GetSpotifyAlbum
	case queue.ResourceType_ARTIST_ID:
		fetchFunc = w.spotifyClient.GetSpotifyArtist
	case queue.ResourceType_ARTIST_NAME:
		// passing through just the artist name, since it doesnt need spotify api call
		err := w.queue.Publish(ctx, models.ScannerQueueName, DomainContent2QueueContent(&models.Content{Artists: []models.ArtistRef{{Name: job.ResourceId}}}, job.Id))
		if err != nil {
			return err
		}
	default:
		return nil
	}

	content, err := fetchFunc(ctx, job.ResourceId)
	if err != nil {
		return err
	}

	err = w.queue.Publish(ctx, models.ScannerQueueName, DomainContent2QueueContent(content, job.Id))
	if err != nil {
		return err
	}
	log.Info().Str("job_id", job.Id).Msg("processed spotify job")

	return nil
}

func DomainContent2QueueContent(c *models.Content, jobId string) *queue.Content {
	var tracks []*queue.Track
	var artists []*queue.ArtistRef

	for _, track := range c.Tracks {
		var artistRefs []*queue.ArtistRef
		for _, a := range track.ArtistRefs {
			artistRefs = append(artistRefs, &queue.ArtistRef{
				Name:       a.Name,
				ExternalId: a.ExternalID,
			})
		}

		tracks = append(tracks, &queue.Track{
			ExternalId: track.ExternalID,
			Name:       track.Name,
			Artists:    artistRefs,
		})
	}

	for _, artist := range c.Artists {
		artists = append(artists, &queue.ArtistRef{
			ExternalId: artist.ExternalID,
			Name:       artist.Name,
		})
	}

	return &queue.Content{
		ScanJobId: jobId,
		Tracks:    tracks,
		Artists:   artists,
	}
}
