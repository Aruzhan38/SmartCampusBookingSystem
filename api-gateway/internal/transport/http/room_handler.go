package http

import (
	"api-gateway/internal/client"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RoomHandler struct {
	roomClient client.RoomClient
}

func NewRoomHandler(roomClient client.RoomClient) *RoomHandler {
	return &RoomHandler{roomClient: roomClient}
}

func (h *RoomHandler) GetRooms(c *gin.Context) {
	resp, err := h.roomClient.GetRooms(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
