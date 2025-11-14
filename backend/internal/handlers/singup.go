package handlers

import (
	"spotiscan/internal/services"

	"net/http"
	"github.com/gin-gonic/gin"
)

func NewSignupHandler(service *services.UserService) *SignupHandler {
	return &SignupHandler{svc: service}
}

type SignupHandler struct {
	svc *services.UserService
}

type Credentials struct{
	Username string `json:"username"`
	Email string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

func (s *SignupHandler) PostSignup(c *gin.Context){
	creds := Credentials{}
	err := c.ShouldBindJSON(&creds)
	if err != nil{
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid credentials format",
			"details": err.Error(),
		})
		return
	}

	token, err := s.svc.Signup(creds.Username, creds.Email, creds.Password)
	switch err{
	case nil:
		c.JSON(http.StatusCreated, gin.H{
			"session_token": token,
		})
	case services.ErrEmailUsed:
		c.JSON(http.StatusConflict, gin.H{
			"error": "Email already in use",
		})
	case services.ErrInvalidEmail:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email format",
		})
	case services.ErrDatabaseFailure:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error",
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
	}
}