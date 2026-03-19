package handlers

import (
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/chivta/spotiscan/internal/appErrors"
	"github.com/chivta/spotiscan/internal/models"
	"github.com/chivta/spotiscan/internal/services"
)

func NewAuthHandler(authService *services.AuthService, validate *validator.Validate) *AuthHandler {
	return &AuthHandler{authService: authService, validate: validate}
}

type AuthHandler struct {
	authService *services.AuthService
	validate    *validator.Validate
}

func (h *AuthHandler) Signup(c *gin.Context) {
	var signupDTO models.SignupDTO
	if err := c.BindJSON(&signupDTO); err != nil {
		RespondWithError(c, appErrors.ErrBadRequest)
		return
	}

	err := h.validate.Struct(signupDTO)
	if err != nil {
		_, ok := err.(validator.ValidationErrors)
		if !ok {
			RespondWithError(c, appErrors.ErrInternal)
			return
		}
		RespondWithError(c, appErrors.ErrBadRequest)
		return
	}

	session, err := h.authService.Signup(c.Request.Context(), signupDTO)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.SetCookie(models.CookieJWT, session.JWT, models.JWTCookieAge, "/", "", false, true)
	c.SetCookie(models.CookieRefreshToken, session.RefreshToken, models.RefreshTokenCookieAge, "/", "", false, true)
	c.JSON(200, gin.H{"message": "signup successful"})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var loginDTO models.LoginDTO
	if err := c.BindJSON(&loginDTO); err != nil {
		RespondWithError(c, appErrors.ErrBadRequest)
		return
	}

	err := h.validate.Struct(loginDTO)
	if err != nil {
		_, ok := err.(validator.ValidationErrors)
		if !ok {
			RespondWithError(c, appErrors.ErrInternal)
			return
		}
		RespondWithError(c, appErrors.ErrBadRequest)
		return
	}

	session, err := h.authService.Login(c.Request.Context(), loginDTO)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.SetCookie(models.CookieJWT, session.JWT, models.JWTCookieAge, "/", "", false, true)
	c.SetCookie(models.CookieRefreshToken, session.RefreshToken, models.RefreshTokenCookieAge, "/", "", false, true)
	c.JSON(200, gin.H{"message": "login successful"})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		RespondWithError(c, appErrors.ErrUnauthorized)
		return
	}

	// TODO: remove
	log.Println(userID)
	userIDInt, err := strconv.Atoi(userID.(string))
	if err != nil {
		RespondWithError(c, appErrors.ErrInternal)
		return
	}

	if err := h.authService.Logout(c.Request.Context(), userIDInt); err != nil {
		RespondWithError(c, err)
		return
	}

	c.SetCookie(models.CookieJWT, "", -1, "/", "", false, true)
	c.SetCookie(models.CookieRefreshToken, "", -1, "/", "", false, true)
	c.JSON(200, gin.H{"message": "logout successful"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	c.JSON(200, gin.H{
		"userID":   userID,
		"userRole": userRole,
	})
}
