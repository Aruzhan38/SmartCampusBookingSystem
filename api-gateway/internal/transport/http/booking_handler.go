package http

import (
	"api-gateway/internal/client"
	"api-gateway/internal/domain"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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

	// extract user_id from JWT claims
	authUserVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	authUser, ok := authUserVal.(*domain.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
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
		uint(authUser.ID),
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
	authUserVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	authUser, ok := authUserVal.(*domain.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userIDParam := c.Query("user_id")
	var userID uint64
	var err error
	if userIDParam == "" {
		if strings.ToUpper(authUser.Role) == "ADMIN" {
			userID = 0
		} else {
			userID = uint64(authUser.ID)
		}
	} else {
		userID, err = parseNonNegativeUintParam(userIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "user_id query parameter must be a non-negative integer",
			})
			return
		}
		if userID == 0 && strings.ToUpper(authUser.Role) != "ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		if userID > 0 && uint(userID) != uint(authUser.ID) && strings.ToUpper(authUser.Role) != "ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
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

	authUserVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	authUser, ok := authUserVal.(*domain.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userIDParam := c.Query("user_id")
	var userID uint64
	var err error
	if userIDParam == "" {
		if strings.ToUpper(authUser.Role) == "ADMIN" {
			userID = 0
		} else {
			userID = uint64(authUser.ID)
		}
	} else {
		userID, err = parseNonNegativeUintParam(userIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "user_id query parameter must be a non-negative integer",
			})
			return
		}
		if userID == 0 && strings.ToUpper(authUser.Role) != "ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		if userID > 0 && uint(userID) != uint(authUser.ID) && strings.ToUpper(authUser.Role) != "ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
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

	// only admins can update booking status
	authUserVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	authUser, ok := authUserVal.(*domain.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if strings.ToUpper(authUser.Role) != "ADMIN" {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
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

func parseNonNegativeUintParam(value string) (uint64, error) {
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, err
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
