package scanner

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/chivta/ruscan/internal/shared/domain"

	"github.com/chivta/ruscan/internal/shared/queue"
)

type jobRepo interface {
	PostJobResult(ctx context.Context, jobID string, status domain.JobStatus, result any) error
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
	if err := w.queue.DeclareQueueWithDLQ(domain.ScannerQueueName); err != nil {
		return err
	}

	deliveries, err := w.queue.ConsumeContent(ctx, domain.ScannerQueueName)
	if err != nil {
		return err
	}

	go w.consumeDead(ctx)

	for delivery := range deliveries {
		job := delivery.Job
		result, err := w.svc.FilterContent(ctx, QueueContent2DomainContent(job))
		if err != nil {
			log.Error().Err(err).Str("job_id", job.ScanJobId).Msg("failed to filter content")
			delivery.Nack(false, true)
			continue
		}
		if err := w.jobRepo.PostJobResult(ctx, job.ScanJobId, domain.JobStatusDone, result); err != nil {
			log.Error().Err(err).Str("job_id", job.ScanJobId).Msg("failed to post job result")
			delivery.Nack(false, true)
			continue
		}
		err = delivery.Ack(false)
		if err != nil {
			log.Error().Err(err).Str("job_id", job.ScanJobId).Msg("failed to ack message")
			return err
		}
	}
	return nil
}

func (w *ScannerWorker) consumeDead(ctx context.Context) {
	deliveries, err := w.queue.ConsumeContent(ctx, domain.ScannerQueueName+".dead")
	if err != nil {
		log.Error().Err(err).Msg("failed to consume scanner dead queue")
		return
	}
	for d := range deliveries {
		if err := w.jobRepo.PostJobResult(ctx, d.Job.ScanJobId, domain.JobStatusFailed, domain.ErrInternal.Code); err != nil {
			log.Error().Err(err).Str("job_id", d.Job.ScanJobId).Msg("failed to mark dead job as failed")
			d.Nack(false, false)
			continue
		}
		err := d.Ack(false)
		if err != nil {
			log.Error().Err(err).Str("job_id", d.Job.ScanJobId).Msg("failed to ack dead job message")
			continue
		}
	}
}

func QueueContent2DomainContent(c *queue.ContentScanJob) *domain.Content {
	tracks := make([]domain.Track, len(c.Tracks))
	for i, t := range c.Tracks {
		var artistRefs []domain.ArtistRef
		for _, a := range t.Artists {
			artistRefs = append(artistRefs, domain.ArtistRef{
				Name:       a.Name,
				ExternalID: a.ExternalId,
			})
		}
		tracks[i] = domain.Track{
			Name:       t.Name,
			ExternalID: t.ExternalId,
			ArtistRefs: artistRefs,
			ImageURL:   t.ImageUrl,
		}
	}

	artists := make([]domain.ArtistRef, len(c.Artists))
	for i, a := range c.Artists {
		artists[i] = domain.ArtistRef{
			Name:       a.Name,
			ExternalID: a.ExternalId,
		}
	}

	return &domain.Content{
		Tracks:  tracks,
		Artists: artists,
	}
}
