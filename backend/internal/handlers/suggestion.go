package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

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
	ArtistName  string `json:"ArtistName" validate:"required"`
	Description string `json:"Description" validate:"required,max=1000"`
}

func (h *SuggestionHandler) CreateArtistInsertSuggestion(c *gin.Context) {
	var req createArtistInsertSuggestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	ArtistName  string `json:"ArtistName" validate:"required"`
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId := c.GetString(models.UserIDKey)
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		RespondWithError(c, err)
		return
	}
	suggestion, err := h.svc.UpdateArtistInsertSuggestion(c.Request.Context(), id, userIdInt, req.ArtistName, req.Description)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, suggestion)
}
