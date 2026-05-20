package http

import (
	"api-gateway/internal/client"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BookingHandler struct {
	bookingClient client.BookingClient
}

func NewBookingHandler(bookingClient client.BookingClient) *BookingHandler {
	return &BookingHandler{bookingClient: bookingClient}
}

type createBookingRequest struct {
	UserID    uint   `json:"user_id" binding:"required,min=1"`
	RoomID    uint   `json:"room_id" binding:"required,min=1"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
	Purpose   string `json:"purpose" binding:"required"`
}

type updateBookingStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

func (h *BookingHandler) CreateBooking(c *gin.Context) {
	var req createBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// TODO: extract user_id from JWT claims instead of trusting the request body.
	if err := validateRFC3339(req.StartTime, "start_time"); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if err := validateRFC3339(req.EndTime, "end_time"); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	resp, err := h.bookingClient.CreateBooking(
		c.Request.Context(),
		req.UserID,
		req.RoomID,
		req.StartTime,
		req.EndTime,
		req.Purpose,
	)
	if err != nil {
		c.JSON(bookingErrorStatus(err), gin.H{
			"error": status.Convert(err).Message(),
		})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *BookingHandler) GetBookingByID(c *gin.Context) {
	id := c.Param("id")
	if _, err := parsePositiveUintParam(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "booking id must be a positive integer",
		})
		return
	}

	resp, err := h.bookingClient.GetBookingByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(bookingErrorStatus(err), gin.H{
			"error": status.Convert(err).Message(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *BookingHandler) ListUserBookings(c *gin.Context) {
	userIDParam := c.Query("user_id")
	userID, err := parsePositiveUintParam(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id query parameter must be a positive integer",
		})
		return
	}

	resp, err := h.bookingClient.ListUserBookings(c.Request.Context(), uint(userID))
	if err != nil {
		c.JSON(bookingErrorStatus(err), gin.H{
			"error": status.Convert(err).Message(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *BookingHandler) CancelBooking(c *gin.Context) {
	id := c.Param("id")
	if _, err := parsePositiveUintParam(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "booking id must be a positive integer",
		})
		return
	}

	userIDParam := c.Query("user_id")
	userID, err := parsePositiveUintParam(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id query parameter must be a positive integer",
		})
		return
	}

	resp, err := h.bookingClient.CancelBooking(c.Request.Context(), id, uint(userID))
	if err != nil {
		c.JSON(bookingErrorStatus(err), gin.H{
			"error": status.Convert(err).Message(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *BookingHandler) UpdateBookingStatus(c *gin.Context) {
	id := c.Param("id")
	if _, err := parsePositiveUintParam(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "booking id must be a positive integer",
		})
		return
	}

	var req updateBookingStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	resp, err := h.bookingClient.UpdateBookingStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		c.JSON(bookingErrorStatus(err), gin.H{
			"error": status.Convert(err).Message(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func validateRFC3339(value string, field string) error {
	if _, err := time.Parse(time.RFC3339, value); err != nil {
		return fmt.Errorf("%s must use RFC3339 format", field)
	}
	return nil
}

func parsePositiveUintParam(value string) (uint64, error) {
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil || parsed == 0 {
		return 0, strconv.ErrSyntax
	}
	return parsed, nil
}

func bookingErrorStatus(err error) int {
	switch status.Code(err) {
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
