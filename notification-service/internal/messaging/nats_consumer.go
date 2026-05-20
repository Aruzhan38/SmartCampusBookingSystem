package messaging

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"notification-service/internal/mail"
	"notification-service/internal/usecase"

	"github.com/nats-io/nats.go"
)

type BookingCreatedEvent struct {
	UserID    uint   `json:"user_id"`
	BookingID uint   `json:"booking_id"`
	RoomID    uint   `json:"room_id"`
	Email     string `json:"email"`
	Message   string `json:"message"`
	Type      string `json:"type"`
}

type BookingStatusChangedEvent struct {
	UserID    uint   `json:"user_id"`
	BookingID uint   `json:"booking_id"`
	RoomID    uint   `json:"room_id"`
	Email     string `json:"email"`
	Message   string `json:"message"`
	Status    string `json:"status"`
	Type      string `json:"type"`
}

type NATSConsumer struct {
	natsURL      string
	usecase      usecase.NotificationUsecase
	mailSender   *mail.SMTPSender
	defaultEmail string
}

func NewNATSConsumer(
	natsURL string,
	uc usecase.NotificationUsecase,
	mailSender *mail.SMTPSender,
	defaultEmail string,
) *NATSConsumer {
	return &NATSConsumer{
		natsURL:      natsURL,
		usecase:      uc,
		mailSender:   mailSender,
		defaultEmail: defaultEmail,
	}
}

func (c *NATSConsumer) Start() error {
	nc, err := nats.Connect(c.natsURL)
	if err != nil {
		return err
	}

	_, err = nc.Subscribe("booking.created", func(msg *nats.Msg) {
		var event BookingCreatedEvent

		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Println("failed to parse booking.created event:", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		notificationType := event.Type
		if notificationType == "" {
			notificationType = "BOOKING_CREATED"
		}

		message := event.Message
		if message == "" {
			message = "Your room booking has been created successfully."
		}

		notification, err := c.usecase.SendNotification(ctx, event.UserID, message, notificationType)
		if err != nil {
			log.Println("failed to save notification:", err)
			return
		}

		emailTo := event.Email
		if emailTo == "" {
			emailTo = c.defaultEmail
		}

		if emailTo != "" {
			err = c.mailSender.Send(
				emailTo,
				"Smart Campus Booking Confirmation",
				message,
			)
			if err != nil {
				log.Println("failed to send email:", err)
			} else {
				log.Println("email sent to:", emailTo)
			}
		}

		log.Println("notification created from NATS event, id:", notification.ID)
	})

	if err != nil {
		return err
	}

	_, err = nc.Subscribe("booking.status_changed", func(msg *nats.Msg) {
		var event BookingStatusChangedEvent

		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Println("failed to parse booking.status_changed event:", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		message := event.Message
		if message == "" {
			message = "Your booking status has been updated."
		}

		notificationType := event.Type
		if notificationType == "" {
			notificationType = "BOOKING_STATUS_CHANGED"
		}

		notification, err := c.usecase.SendNotification(ctx, event.UserID, message, notificationType)
		if err != nil {
			log.Println("failed to save notification:", err)
			return
		}

		emailTo := event.Email
		if emailTo == "" {
			log.Println("booking.status_changed event missing user email; email will not be sent")
			return
		}

		title := "Smart Campus Booking Update"
		err = c.mailSender.Send(
			emailTo,
			title,
			message,
		)
		if err != nil {
			log.Println("failed to send email:", err)
		} else {
			log.Println("status change email sent to:", emailTo)
		}

		log.Println("notification created from booking.status_changed event, id:", notification.ID)
	})

	if err != nil {
		return err
	}

	log.Println("Notification Service subscribed to NATS subject: booking.created")
	return nil
}
