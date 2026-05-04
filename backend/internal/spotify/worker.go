package spotify

import (
	"context"

	"github.com/chivta/ruscan/internal/shared/appErrors"
	"github.com/chivta/ruscan/internal/shared/models"
	"github.com/chivta/ruscan/internal/shared/queue"
	"github.com/rs/zerolog/log"
)

type JobRepo interface {
	PostJobResult(ctx context.Context, jobID string, status models.JobStatus, result any) error
}

func NewSpotifyGatewayWorker(jobRepo JobRepo, queue *queue.Client, spotifyClient *SpotifyClient) *SpotifyGatewayWorker {
	return &SpotifyGatewayWorker{jobRepo: jobRepo, queue: queue, spotifyClient: spotifyClient}
}

type SpotifyGatewayWorker struct {
	jobRepo       JobRepo
	queue         *queue.Client
	spotifyClient *SpotifyClient
}

func (w *SpotifyGatewayWorker) Start(ctx context.Context) error {
	err := w.queue.DeclareQueue(models.SpotifyQueueName)
	if err != nil {
		return err
	}

	deliveries, err := w.queue.ConsumeContentFetchJobs(ctx, models.SpotifyQueueName)
	if err != nil {
		return err
	}

	for delivery := range deliveries {
		ack := w.processJob(ctx, delivery.Job)
		if ack {
			delivery.Msg.Ack(false)
		} else {
			delivery.Msg.Nack(false, true)
		}
	}
	return nil
}

func (w *SpotifyGatewayWorker) processJob(ctx context.Context, job *queue.ContentFetchJob) bool {
	log.Info().Str("job_id", job.Id).Str("resource_type", string(job.ResourceType)).Str("resource_id", job.ResourceId).Msg("processing spotify gateway job")
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
			log.Error().Err(err).Msg("failed to publish content for artist name job")
			return false
		}
		return true
	default:
		log.Error().Str("resource_type", string(job.ResourceType)).Msg("unknown resource type")
		return true
	}
	content, err := fetchFunc(ctx, job.ResourceId)
	if err != nil {
		log.Error().Err(err).Str("resource_id", job.ResourceId).Str("resource_type", string(job.ResourceType)).Msg("failed to fetch content from spotify")
		spotifyErr, ok := err.(*SpotifyError)
		if ok {
			switch spotifyErr.Status {
			case 404:
				err = appErrors.ErrSpotifyNotFound
			case 400:
				err = appErrors.ErrBadRequest
			case 429:
				err = appErrors.ErrTooManyRequests
			default:
				log.Error().Err(spotifyErr).Int("status", spotifyErr.Status).Msg("spotify API error")
				err = appErrors.ErrSpotifyAPIError
			}
			log.Error().Err(err).Str("resource_id", job.ResourceId).Str("resource_type", string(job.ResourceType)).Msg("spotify API error during fetching")
			err = w.jobRepo.PostJobResult(ctx, job.Id, models.JobStatusFailed, err.Error())
			if err != nil {
				log.Error().Err(err).Msg("failed to post job result for spotify error")
				return false
			}
			return true
		} else {
			log.Error().Err(err).Msg("uknown error during spotify fetching")
			return false
		}
	}

	err = w.queue.Publish(ctx, models.ScannerQueueName, DomainContent2QueueContent(content, job.Id))
	if err != nil {
		log.Error().Err(err).Msg("failed to publish content for spotify job")
		return false
	}
	log.Info().Str("job_id", job.Id).Msg("processed spotify job") // TODO: remove ts

	return true
}

func DomainContent2QueueContent(c *models.Content, jobId string) *queue.ContentScanJob {
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
			ImageUrl:   track.ImageURL,
		})
	}

	for _, artist := range c.Artists {
		artists = append(artists, &queue.ArtistRef{
			ExternalId: artist.ExternalID,
			Name:       artist.Name,
		})
	}

	return &queue.ContentScanJob{
		ScanJobId: jobId,
		Tracks:    tracks,
		Artists:   artists,
	}
}
