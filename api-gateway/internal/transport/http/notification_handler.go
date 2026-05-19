package http

import (
	"api-gateway/internal/client"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"
)

type NotificationHandler struct {
	notificationClient client.NotificationClient
}

func NewNotificationHandler(notificationClient client.NotificationClient) *NotificationHandler {
	return &NotificationHandler{notificationClient: notificationClient}
}

func (h *NotificationHandler) ListMyNotifications(c *gin.Context) {
	user, ok := getUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if !isAdmin(user) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can view notifications"})
		return
	}

	resp, err := h.notificationClient.ListUserNotifications(c.Request.Context(), uint(user.ID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": status.Convert(err).Message()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *NotificationHandler) GetNotificationByID(c *gin.Context) {
	user, ok := getUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if !isAdmin(user) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can view notifications"})
		return
	}

	id := c.Param("id")
	resp, err := h.notificationClient.GetNotificationByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": status.Convert(err).Message()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	user, ok := getUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if !isAdmin(user) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can update notification status"})
		return
	}

	id := c.Param("id")
	resp, err := h.notificationClient.MarkAsRead(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": status.Convert(err).Message()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
