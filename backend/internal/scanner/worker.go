package scanner

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/chivta/ruscan/internal/shared/appErrors"
	"github.com/chivta/ruscan/internal/shared/models"
	"github.com/chivta/ruscan/internal/shared/queue"
)

type jobRepo interface {
	PostJobResult(ctx context.Context, jobID string, status models.JobStatus, result any) error
}

func NewScannerWorker(q *queue.Client, svc *SpotifyService, jobRepo jobRepo) *ScannerWorker {
	return &ScannerWorker{queue: q, svc: svc, jobRepo: jobRepo}
}

type ScannerWorker struct {
	queue   *queue.Client
	svc     *SpotifyService
	jobRepo jobRepo
}

func (w *ScannerWorker) Start(ctx context.Context) error {
	if err := w.queue.DeclareQueueWithDLQ(models.ScannerQueueName); err != nil {
		return err
	}

	deliveries, err := w.queue.ConsumeContent(ctx, models.ScannerQueueName)
	if err != nil {
		return err
	}

	go w.consumeDead(ctx)

	for delivery := range deliveries {
		job := delivery.Job
		result, err := w.svc.FilterContent(ctx, QueueContent2DomainContent(job))
		if err != nil {
			log.Error().Err(err).Str("job_id", job.ScanJobId).Msg("failed to filter content")
			delivery.Msg.Nack(false, true)
			continue
		}
		if err := w.jobRepo.PostJobResult(ctx, job.ScanJobId, models.JobStatusDone, result); err != nil {
			log.Error().Err(err).Str("job_id", job.ScanJobId).Msg("failed to post job result")
			delivery.Msg.Nack(false, true)
			continue
		}
		err = delivery.Msg.Ack(false)
		if err != nil {
			log.Error().Err(err).Str("job_id", job.ScanJobId).Msg("failed to ack message")
			return err
		}
	}
	return nil
}

func (w *ScannerWorker) consumeDead(ctx context.Context) {
	deliveries, err := w.queue.ConsumeContent(ctx, models.ScannerQueueName+".dead")
	if err != nil {
		log.Error().Err(err).Msg("failed to consume scanner dead queue")
		return
	}
	for d := range deliveries {
		if err := w.jobRepo.PostJobResult(ctx, d.Job.ScanJobId, models.JobStatusFailed, appErrors.ErrInternal.Code); err != nil {
			log.Error().Err(err).Str("job_id", d.Job.ScanJobId).Msg("failed to mark dead job as failed")
			d.Msg.Nack(false, false)
			continue
		}
		d.Msg.Ack(false)
	}
}

func QueueContent2DomainContent(c *queue.ContentScanJob) *models.Content {
	tracks := make([]models.Track, len(c.Tracks))
	for i, t := range c.Tracks {
		var artistRefs []models.ArtistRef
		for _, a := range t.Artists {
			artistRefs = append(artistRefs, models.ArtistRef{
				Name:       a.Name,
				ExternalID: a.ExternalId,
			})
		}
		tracks[i] = models.Track{
			Name:       t.Name,
			ExternalID: t.ExternalId,
			ArtistRefs: artistRefs,
			ImageURL:   t.ImageUrl,
		}
	}

	artists := make([]models.ArtistRef, 0, len(c.Artists))
	for i, a := range c.Artists {
		artists[i] = models.ArtistRef{
			Name:       a.Name,
			ExternalID: a.ExternalId,
		}
	}

	return &models.Content{
		Tracks:  tracks,
		Artists: artists,
	}
}
