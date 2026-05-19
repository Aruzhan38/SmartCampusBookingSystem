package http

import (
	"api-gateway/internal/client"
	"api-gateway/internal/domain"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userClient client.UserClient
}

func NewUserHandler(userClient client.UserClient) *UserHandler {
	return &UserHandler{userClient: userClient}
}

type registerRequest struct {
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *UserHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate email domain
	if !isValidStudentEmail(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only @gmail.com email addresses are allowed"})
		return
	}

	if err := h.userClient.RegisterUser(c.Request.Context(), req.FullName, req.Email, req.Password, req.Role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Registration successful"})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.userClient.LoginUser(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *UserHandler) Profile(c *gin.Context) {
	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, ok := userValue.(*domain.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user context"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        user.ID,
		"full_name": user.FullName,
		"email":     user.Email,
		"role":      user.Role,
	})
}

func isValidStudentEmail(email string) bool {
	t := strings.ToLower(email)
	return strings.HasSuffix(t, "@gmail.com")
}
