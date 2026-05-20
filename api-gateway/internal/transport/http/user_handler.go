package http

import (
	"api-gateway/internal/client"
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

	if err := h.userClient.RegisterUser(c.Request.Context(), req.FullName, req.Email, req.Password, req.Role); err != nil {
		errMsg := "Registration failed"
		if strings.Contains(strings.ToLower(err.Error()), "user already exists") {
			errMsg = "User already exists"
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
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
		errMsg := "Invalid credentials"
		if !strings.Contains(strings.ToLower(err.Error()), "invalid credentials") {
			errMsg = err.Error()
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": errMsg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id is required"})
		return
	}

	user, err := h.userClient.GetUserByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}
