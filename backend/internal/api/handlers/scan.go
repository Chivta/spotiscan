package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/chivta/ruscan/internal/shared/models"
	"github.com/chivta/ruscan/internal/shared/queue"
)

type jobRepo interface {
	CreateJob(ctx context.Context, jobID string) error
	GetJob(ctx context.Context, jobID string) (*models.Job, error)
}

func NewScanHandler(jobRepo jobRepo, queueClient *queue.Client, providers map[string]struct{}, validate *validator.Validate) *ScanHandler {
	return &ScanHandler{jobRepo: jobRepo, q: queueClient, providers: providers, validate: validate}
}

type ScanHandler struct {
	jobRepo   jobRepo
	q         *queue.Client
	providers map[string]struct{}
	validate  *validator.Validate
}

func (h *ScanHandler) enqueue(c *gin.Context, resourceType queue.ResourceType, resourceID string) {
	queueName := c.Param("provider")
	if _, ok := h.providers[queueName]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider"}) // TODO: use sentinel
		return
	}

	jobID := uuid.New().String()
	if err := h.jobRepo.CreateJob(c.Request.Context(), jobID); err != nil {
		RespondWithError(c, err)
		return
	}

	if err := h.q.Publish(c.Request.Context(), queueName, &queue.ScanJob{
		Id:           jobID,
		UserId:       c.GetString(models.UserIDKey),
		ResourceType: resourceType,
		ResourceId:   resourceID,
	}); err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"jobId": jobID})
}

func (h *ScanHandler) ScanPlaylist(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty id"})
		return
	}
	h.enqueue(c, queue.ResourceType_PLAYLIST_ID, id)
}

func (h *ScanHandler) ScanTrack(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty id"})
		return
	}
	h.enqueue(c, queue.ResourceType_TRACK_ID, id)
}

func (h *ScanHandler) ScanAlbum(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty id"})
		return
	}
	h.enqueue(c, queue.ResourceType_ALBUM_ID, id)
}

func (h *ScanHandler) ScanArtist(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty id"})
		return
	}
	h.enqueue(c, queue.ResourceType_ARTIST_ID, id)
}

func (h *ScanHandler) ScanArtistByName(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty name"})
		return
	}
	h.enqueue(c, queue.ResourceType_ARTIST_NAME, name)
}

func (h *ScanHandler) GetJobStatus(c *gin.Context) {
	jobID := c.Param("jobId")
	job, err := h.jobRepo.GetJob(c.Request.Context(), jobID)
	if err != nil {
		RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, job)
}
