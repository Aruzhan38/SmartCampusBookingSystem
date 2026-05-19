package http

import (
	"api-gateway/internal/client"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BookingHandler struct {
	bookingClient      client.BookingClient
	notificationClient client.NotificationClient
	userClient         client.UserClient
}

func NewBookingHandler(bookingClient client.BookingClient, notificationClient client.NotificationClient, userClient client.UserClient) *BookingHandler {
	return &BookingHandler{bookingClient: bookingClient, notificationClient: notificationClient, userClient: userClient}
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
	user, ok := getUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req createBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
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

	if !isAdmin(user) {
		req.UserID = uint(user.ID)
	}

	resp, err := h.bookingClient.CreateBooking(
		c.Request.Context(),
		req.UserID,
		req.RoomID,
		req.StartTime,
		req.EndTime,
		req.Purpose,
		user.Role,
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
	user, ok := getUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
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

	if !isAdmin(user) && uint(userID) != uint(user.ID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can query other users' bookings"})
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

func (h *BookingHandler) ListAllBookings(c *gin.Context) {
	user, ok := getUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if !isAdmin(user) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can view all bookings"})
		return
	}

	resp, err := h.bookingClient.ListUserBookings(c.Request.Context(), 0)
	if err != nil {
		c.JSON(bookingErrorStatus(err), gin.H{
			"error": status.Convert(err).Message(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *BookingHandler) CancelBooking(c *gin.Context) {
	user, ok := getUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

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

	if !isAdmin(user) && uint(userID) != uint(user.ID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can cancel other users' bookings"})
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
	user, ok := getUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if !isAdmin(user) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can update booking status"})
		return
	}

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

	if strings.EqualFold(req.Status, "Confirmed") || strings.EqualFold(req.Status, "Rejected") || strings.EqualFold(req.Status, "Cancelled") {
		bookingUserID := resp.Booking.GetUserId()
		bookingUser, err := h.userClient.GetUserByID(c.Request.Context(), bookingUserID)
		if err != nil {
			log.Printf("failed to resolve booking owner email: %v", err)
		} else if bookingUser.Email != "" {
			notificationType := "booking_confirmed"
			emailBody := fmt.Sprintf("Hello %s,\n\nYour booking for room %s from %s to %s has been approved.\n\nThank you,\nSmart Campus Booking System", bookingUser.FullName, resp.Booking.RoomId, resp.Booking.StartTime, resp.Booking.EndTime)
			if strings.EqualFold(req.Status, "Rejected") {
				notificationType = "booking_rejected"
				emailBody = fmt.Sprintf("Hello %s,\n\nYour booking for room %s from %s to %s has been rejected.\n\nPlease contact the administrator for details.\nSmart Campus Booking System", bookingUser.FullName, resp.Booking.RoomId, resp.Booking.StartTime, resp.Booking.EndTime)
			} else if strings.EqualFold(req.Status, "Cancelled") {
				notificationType = "booking_cancelled"
				emailBody = fmt.Sprintf("Hello %s,\n\nYour booking for room %s from %s to %s has been cancelled.\n\nThank you,\nSmart Campus Booking System", bookingUser.FullName, resp.Booking.RoomId, resp.Booking.StartTime, resp.Booking.EndTime)
			}

			if _, err := h.notificationClient.SendNotification(
				c.Request.Context(),
				bookingUser.Email,
				emailBody,
				notificationType,
			); err != nil {
				log.Printf("failed to send notification email: %v", err)
			} else {
				log.Printf("notification email sent to %s for status %s", bookingUser.Email, req.Status)
			}
		} else {
			log.Printf("booking owner %s has no email address", bookingUserID)
		}
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
