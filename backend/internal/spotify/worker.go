package spotify

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/chivta/ruscan/internal/shared/domain"
	"github.com/chivta/ruscan/internal/shared/queue"
)

type JobRepo interface {
	PostJobResult(ctx context.Context, jobID string, status domain.JobStatus, result any) error
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
	if err := w.queue.DeclareQueueWithDLQ(domain.SpotifyQueueName); err != nil {
		return err
	}

	deliveries, err := w.queue.ConsumeContentFetchJobs(ctx, domain.SpotifyQueueName)
	if err != nil {
		return err
	}

	go w.consumeDead(ctx)

	for delivery := range deliveries {
		ack := w.processJob(ctx, delivery.Job)
		if ack {
			err = delivery.Msg.Ack(false)
			if err != nil {
				log.Error().Err(err).Str("job_id", delivery.Job.Id).Msg("failed to ack message")
				return err
			}
		} else {
			err = delivery.Msg.Nack(false, true)
			if err != nil {
				log.Error().Err(err).Str("job_id", delivery.Job.Id).Msg("failed to nack message")
				return err
			}
		}
	}
	return nil
}

func (w *SpotifyGatewayWorker) consumeDead(ctx context.Context) {
	deliveries, err := w.queue.ConsumeContentFetchJobs(ctx, domain.SpotifyQueueName+".dead")
	if err != nil {
		log.Error().Err(err).Msg("failed to consume spotify dead queue")
		return
	}
	for d := range deliveries {
		if err := w.jobRepo.PostJobResult(ctx, d.Job.Id, domain.JobStatusFailed, domain.ErrInternal.Code); err != nil {
			log.Error().Err(err).Str("job_id", d.Job.Id).Msg("failed to mark dead job as failed")
			d.Msg.Nack(false, false)
			continue
		}
		d.Msg.Ack(false)
	}
}

func (w *SpotifyGatewayWorker) processJob(ctx context.Context, job *queue.ContentFetchJob) bool {
	var fetchFunc func(context.Context, string) (*domain.Content, error)
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
		err := w.queue.Publish(ctx, domain.ScannerQueueName, DomainContent2QueueContent(&domain.Content{Artists: []domain.ArtistRef{{Name: job.ResourceId}}}, job.Id))
		if err != nil {
			log.Error().Err(err).Msg("failed to publish content for artist name job")
			return false
		}
		return true
	default:
		log.Error().Str("resource_type", string(job.ResourceType)).Msg("unknown resource type")
		w.jobRepo.PostJobResult(ctx, job.Id, domain.JobStatusFailed, domain.ErrBadRequest.Code)
		return true
	}
	content, err := fetchFunc(ctx, job.ResourceId)
	if err != nil {
		spotifyErr, ok := err.(*SpotifyError)
		if ok {
			switch spotifyErr.Status {
			case 404:
				err = domain.ErrSpotifyNotFound
			case 400:
				err = domain.ErrBadRequest
			case 429:
				err = domain.ErrTooManyRequests
			default:
				log.Error().Err(spotifyErr).Int("status", spotifyErr.Status).Msg("spotify API error")
				err = domain.ErrSpotifyAPIError
			}
			err = w.jobRepo.PostJobResult(ctx, job.Id, domain.JobStatusFailed, err.Error())
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

	err = w.queue.Publish(ctx, domain.ScannerQueueName, DomainContent2QueueContent(content, job.Id))
	if err != nil {
		log.Error().Err(err).Msg("failed to publish content for spotify job")
		return false
	}

	return true
}

func DomainContent2QueueContent(c *domain.Content, jobId string) *queue.ContentScanJob {
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
