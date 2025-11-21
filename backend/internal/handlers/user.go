package handlers

import (
	"net/http"
	"spotiscan/internal/services"

	"github.com/gin-gonic/gin"
)

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{svc: service}
}

type UserHandler struct {
	svc *services.UserService
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userId, exists := c.Get("user_id")
	if ! exists{
		c.Status(http.StatusUnauthorized)
	}
	user, err := h.svc.GetUser(userId.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}
	c.JSON(http.StatusOK, user)
}