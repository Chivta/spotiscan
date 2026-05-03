package scanner

import (
	"context"

	"github.com/chivta/ruscan/internal/shared/models"
	"github.com/chivta/ruscan/internal/shared/queue"
)

func NewScannerWorker(queue *queue.Client, svc *SpotifyService) *ScannerWorker {
	return &ScannerWorker{queue: queue, svc: svc}
}

type ScannerWorker struct {
	queue *queue.Client
	svc   *SpotifyService
}

func (w *ScannerWorker) Start(ctx context.Context) error {
	err := w.queue.DeclareQueue(models.ScannerQueueName)
	if err != nil {
		return err
	}

	deliveries, err := w.queue.ConsumeContent(ctx, models.ScannerQueueName)
	if err != nil {
		return err
	}

	for delivery := range deliveries {
		job := delivery.Job
		if ok := w.svc.ProcessContentScanJob(ctx, QueueContent2DomainContent(job), job.ScanJobId); ok {
			delivery.Msg.Ack(false)
		} else {
			delivery.Msg.Nack(false, true)
		}
	}
	return nil
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
		}
	}

	artists := make([]models.ArtistRef, len(c.Artists))
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
