package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/chivta/ruscan/internal/appErrors"
	"github.com/chivta/ruscan/internal/models"
	"github.com/chivta/ruscan/internal/services"
)

func NewSuggestionHandler(service *services.SuggestionService, validate *validator.Validate) *SuggestionHandler {
	return &SuggestionHandler{svc: service, validate: validate}
}

type SuggestionHandler struct {
	svc      *services.SuggestionService
	validate *validator.Validate
}

type createArtistInsertSuggestionRequest struct {
	ArtistName  string `json:"ArtistName" validate:"required,max=255"`
	Description string `json:"Description" validate:"required,max=1000"`
}

func (h *SuggestionHandler) CreateArtistInsertSuggestion(c *gin.Context) {
	var req createArtistInsertSuggestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		RespondWithError(c,appErrors.ErrBadRequest)
		return
	}

	userId := c.GetString(models.UserIDKey)
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		RespondWithError(c, err)
		return
	}
	suggestion, err := h.svc.CreateArtistInsertSuggestion(c.Request.Context(), req.ArtistName, req.Description, userIdInt)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, suggestion)
}

func (h *SuggestionHandler) GetArtistInsertSuggestions(c *gin.Context) {
	userId := c.GetString(models.UserIDKey)
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		RespondWithError(c, err)
		return
	}
	suggestions, err := h.svc.GetArtistInsertSuggestions(c.Request.Context(), userIdInt)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, suggestions)
}

func (h *SuggestionHandler) DeleteArtistInsertSuggestion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid suggestion id"})
		return
	}

	userId := c.GetString(models.UserIDKey)
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		RespondWithError(c, err)
		return
	}
	if err := h.svc.DeleteArtistInsertSuggestion(c.Request.Context(), id, userIdInt); err != nil {
		RespondWithError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

type updateArtistInsertSuggestionRequest struct {
	Description string `json:"Description" validate:"required,max=1000"`
}

func (h *SuggestionHandler) UpdateArtistInsertSuggestion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid suggestion id"})
		return
	}

	var req updateArtistInsertSuggestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		RespondWithError(c,appErrors.ErrBadRequest)
		return
	}

	userId := c.GetString(models.UserIDKey)
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		RespondWithError(c, err)
		return
	}
	suggestion, err := h.svc.UpdateArtistInsertSuggestion(c.Request.Context(), id, userIdInt, req.Description)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, suggestion)
}

type createArtistDeleteSuggestionRequest struct {
	ArtistName  string `json:"ArtistName" validate:"required,max=255"`
	Description string `json:"Description" validate:"required,max=1000"`
}

func (h *SuggestionHandler) CreateArtistDeleteSuggestion(c *gin.Context) {
	var req createArtistDeleteSuggestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		RespondWithError(c,appErrors.ErrBadRequest)
		return
	}

	userId := c.GetString(models.UserIDKey)
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		RespondWithError(c, err)
		return
	}
	suggestion, err := h.svc.CreateArtistDeleteSuggestion(c.Request.Context(), req.ArtistName, req.Description, userIdInt)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, suggestion)
}

func (h *SuggestionHandler) GetArtistDeleteSuggestions(c *gin.Context) {
	userId := c.GetString(models.UserIDKey)
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		RespondWithError(c, err)
		return
	}
	suggestions, err := h.svc.GetArtistDeleteSuggestions(c.Request.Context(), userIdInt)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, suggestions)
}

func (h *SuggestionHandler) DeleteArtistDeleteSuggestion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid suggestion id"})
		return
	}

	userId := c.GetString(models.UserIDKey)
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		RespondWithError(c, err)
		return
	}
	if err := h.svc.DeleteArtistDeleteSuggestion(c.Request.Context(), id, userIdInt); err != nil {
		RespondWithError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

type updateArtistDeleteSuggestionRequest struct {
	Description string `json:"Description" validate:"required,max=1000"`
}

func (h *SuggestionHandler) UpdateArtistDeleteSuggestion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid suggestion id"})
		return
	}

	var req updateArtistDeleteSuggestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		RespondWithError(c,appErrors.ErrBadRequest)
		return
	}

	userId := c.GetString(models.UserIDKey)
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		RespondWithError(c, err)
		return
	}
	suggestion, err := h.svc.UpdateArtistDeleteSuggestion(c.Request.Context(), id, userIdInt, req.Description)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, suggestion)
}

func (h *SuggestionHandler) GetAllArtistInsertSuggestions(c *gin.Context) {
	suggestions, err := h.svc.GetAllArtistInsertSuggestions(c.Request.Context())
	if err != nil {
		RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, suggestions)
}

func (h *SuggestionHandler) ApproveArtistInsertSuggestion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid suggestion id"})
		return
	}
	suggestion, err := h.svc.ApproveArtistInsertSuggestion(c.Request.Context(), id)
	if err != nil {
		RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, suggestion)
}

func (h *SuggestionHandler) GetAllArtistDeleteSuggestions(c *gin.Context) {
	suggestions, err := h.svc.GetAllArtistDeleteSuggestions(c.Request.Context())
	if err != nil {
		RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, suggestions)
}

func (h *SuggestionHandler) ApproveArtistDeleteSuggestion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid suggestion id"})
		return
	}
	suggestion, err := h.svc.ApproveArtistDeleteSuggestion(c.Request.Context(), id)
	if err != nil {
		RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, suggestion)
}

type declineSuggestionRequest struct {
	DeclineReason string `json:"DeclineReason" validate:"required,max=1000"`
}

func (h *SuggestionHandler) DeclineArtistInsertSuggestion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid suggestion id"})
		return
	}

	var req declineSuggestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if err := h.validate.Struct(req); err != nil {
		RespondWithError(c, appErrors.ErrBadRequest)
		return
	}

	suggestion, err := h.svc.DeclineArtistInsertSuggestion(c.Request.Context(), id, req.DeclineReason)
	if err != nil {
		RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, suggestion)
}

func (h *SuggestionHandler) DeclineArtistDeleteSuggestion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid suggestion id"})
		return
	}

	var req declineSuggestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if err := h.validate.Struct(req); err != nil {
		RespondWithError(c, appErrors.ErrBadRequest)
		return
	}

	suggestion, err := h.svc.DeclineArtistDeleteSuggestion(c.Request.Context(), id, req.DeclineReason)
	if err != nil {
		RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, suggestion)
}
