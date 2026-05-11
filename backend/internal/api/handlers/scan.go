package handlers

import (
	"context"
	"net/http"

	"github.com/chivta/ruscan/internal/shared/domain"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/chivta/ruscan/internal/shared/queue"
)

type jobRepo interface {
	CreateJob(ctx context.Context, jobID, userID string) error
	GetJob(ctx context.Context, jobID, userID string) (*domain.Job, error)
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

type spotifyIDParam struct {
	ID string `validate:"required,alphanum,len=22"`
}

type artistNameParam struct {
	Name string `validate:"required,min=1,max=255"`
}

func (h *ScanHandler) validateID(c *gin.Context, id string) bool {
	if err := h.validate.Struct(spotifyIDParam{ID: id}); err != nil {
		RespondWithError(c, domain.ErrBadRequest)
		return false
	}
	return true
}

func (h *ScanHandler) validateName(c *gin.Context, name string) bool {
	if err := h.validate.Struct(artistNameParam{Name: name}); err != nil {
		RespondWithError(c, domain.ErrBadRequest)
		return false
	}
	return true
}

func (h *ScanHandler) enqueue(c *gin.Context, resourceType queue.ResourceType, resourceID string) {
	userID := c.GetString(domain.UserIDKey)
	queueName := c.Param("provider")
	if _, ok := h.providers[queueName]; !ok {
		RespondWithError(c, domain.ErrBadRequest)
		return
	}

	jobID := uuid.New().String()
	if err := h.jobRepo.CreateJob(c.Request.Context(), jobID, userID); err != nil {
		RespondWithError(c, err)
		return
	}

	if err := h.q.Publish(c.Request.Context(), queueName, &queue.ContentFetchJob{
		Id:           jobID,
		UserId:       c.GetString(domain.UserIDKey),
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
	if !h.validateID(c, id) {
		return
	}
	h.enqueue(c, queue.ResourceType_PLAYLIST_ID, id)
}

func (h *ScanHandler) ScanTrack(c *gin.Context) {
	id := c.Param("id")
	if !h.validateID(c, id) {
		return
	}
	h.enqueue(c, queue.ResourceType_TRACK_ID, id)
}

func (h *ScanHandler) ScanAlbum(c *gin.Context) {
	id := c.Param("id")
	if !h.validateID(c, id) {
		return
	}
	h.enqueue(c, queue.ResourceType_ALBUM_ID, id)
}

func (h *ScanHandler) ScanArtist(c *gin.Context) {
	id := c.Param("id")
	if !h.validateID(c, id) {
		return
	}
	h.enqueue(c, queue.ResourceType_ARTIST_ID, id)
}

func (h *ScanHandler) ScanArtistByName(c *gin.Context) {
	name := c.Query("name")
	if !h.validateName(c, name) {
		return
	}
	h.enqueue(c, queue.ResourceType_ARTIST_NAME, name)
}

func (h *ScanHandler) GetJobStatus(c *gin.Context) {
	userID := c.GetString(domain.UserIDKey)
	jobID := c.Param("jobId")
	job, err := h.jobRepo.GetJob(c.Request.Context(), jobID, userID)
	if err != nil {
		RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, job)
}
