package http

import (
	"api-gateway/internal/client"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RoomHandler struct {
	roomClient client.RoomClient
}

func NewRoomHandler(roomClient client.RoomClient) *RoomHandler {
	return &RoomHandler{roomClient: roomClient}
}

type createRoomRequest struct {
	RoomNumber  string `json:"room_number" binding:"required"`
	Capacity    int32  `json:"capacity" binding:"required,min=1"`
	BuildingID  string `json:"building_id" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type updateRoomRequest struct {
	RoomNumber  string `json:"room_number" binding:"required"`
	Capacity    int32  `json:"capacity" binding:"required,min=1"`
	BuildingID  string `json:"building_id" binding:"required"`
	Description string `json:"description" binding:"required"`
}

func (h *RoomHandler) GetRooms(c *gin.Context) {
	resp, err := h.roomClient.GetRooms(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *RoomHandler) GetRoomByID(c *gin.Context) {
	id := c.Param("id")
	if _, err := parsePositiveUint(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "room id must be a positive integer",
		})
		return
	}

	resp, err := h.roomClient.GetRoomByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(roomErrorStatus(err), gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *RoomHandler) CreateRoom(c *gin.Context) {
	var req createRoomRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	resp, err := h.roomClient.CreateRoom(
		c.Request.Context(),
		req.RoomNumber,
		req.Capacity,
		req.BuildingID,
		req.Description,
	)

	if err != nil {
		c.JSON(roomErrorStatus(err), gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *RoomHandler) UpdateRoom(c *gin.Context) {
	id := c.Param("id")
	if _, err := parsePositiveUint(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "room id must be a positive integer",
		})
		return
	}

	var req updateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	resp, err := h.roomClient.UpdateRoom(
		c.Request.Context(),
		id,
		req.RoomNumber,
		req.Capacity,
		req.BuildingID,
		req.Description,
	)
	if err != nil {
		c.JSON(roomErrorStatus(err), gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *RoomHandler) SearchRoomsByCapacity(c *gin.Context) {
	capacityParam := c.Query("capacity")
	if capacityParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "capacity query parameter is required",
		})
		return
	}

	capacity, err := strconv.ParseInt(capacityParam, 10, 32)
	if err != nil || capacity < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "capacity must be a positive integer",
		})
		return
	}

	resp, err := h.roomClient.SearchRoomsByCapacity(c.Request.Context(), int32(capacity))
	if err != nil {
		c.JSON(roomErrorStatus(err), gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func roomErrorStatus(err error) int {
	switch status.Code(err) {
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.NotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func parsePositiveUint(value string) (uint64, error) {
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil || parsed == 0 {
		return 0, strconv.ErrSyntax
	}
	return parsed, nil
}
