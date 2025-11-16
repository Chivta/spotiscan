package handlers

import (
	
	"net/http"
	"spotiscan/internal/services"

	"github.com/gin-gonic/gin"
)

func NewAuthHandler(service *services.AuthService) *AuthHandler {
	return &AuthHandler{svc: service}
}

type AuthHandler struct {
	svc *services.AuthService
}

type signupCredentials struct {
	Username string `json:"username" binding:"required,min=3,max=30,ascii"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=70"`
}

func (s *AuthHandler) PostSignup(c *gin.Context) {
	creds := signupCredentials{}
	err := c.ShouldBindJSON(&creds)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid credentials format",
			"details": err.Error(),
		})
		return
	}

	token, err := s.svc.Signup(creds.Username, creds.Email, creds.Password)
	switch err {
	case nil:
		c.SetCookie("session_token", token, 3600*24*7, "/", "", true, true)
		c.Status(http.StatusCreated)
		return
	case services.ErrEmailUsed:
		c.JSON(http.StatusConflict, gin.H{
			"error": "Email already in use",
			"code":  "EMAIL_IN_USE",
		})
	case services.ErrUsernameUsed:
		c.JSON(http.StatusConflict, gin.H{
			"error": "Username already in use",
			"code":  "USERNAME_IN_USE",
		})
	case services.ErrInvalidEmail:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email format",
			"code":  "INVALID_EMAIL",
		})
	case services.ErrDatabaseFailure:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error",
			"code":  "DATABASE_ERROR",
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
			"code":  "INTERNAL_ERROR",
		})
	}
}

func (s *AuthHandler) PostLogout(c *gin.Context) {
	sessionToken, err := c.Cookie("session_token")
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}
	s.svc.Logout(sessionToken)
	c.SetCookie("session_token", "", -1, "/", "", true, true)
	c.Status(http.StatusNoContent)
}

type loginCredentials struct {
	EmailOrUsername string `json:"emailOrUsername" binding:"required"`
	Password string `json:"password" binding:"required,min=8,max=70"`
}

func (s *AuthHandler) PostLogin(c *gin.Context) {
	creds := loginCredentials{}
	err := c.ShouldBindJSON(&creds)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid credentials format",
			"details": err.Error(),
		})
		return
	}

	token, err := s.svc.Login(creds.EmailOrUsername, creds.Password)
	switch err {
	case nil:
		c.SetCookie("session_token", token, 3600*24*7, "/", "", true, true)
		c.Status(http.StatusOK)
		return
	case services.ErrInvalidCredentials:
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email/username or password",
			"code":  "INVALID_CREDENTIALS",
		})
	case services.ErrDatabaseFailure:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error",
			"code":  "DATABASE_ERROR",
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
			"code":  "INTERNAL_ERROR",
		})
	}
}