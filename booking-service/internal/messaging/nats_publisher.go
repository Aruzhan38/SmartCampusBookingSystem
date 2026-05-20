package messaging

import (
	"encoding/json"

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

type NATSPublisher interface {
	PublishBookingCreated(event BookingCreatedEvent) error
	PublishBookingStatusChanged(event BookingStatusChangedEvent) error
}

type natsPublisher struct {
	conn *nats.Conn
}

func NewNATSPublisher(natsURL string) (NATSPublisher, error) {
	conn, err := nats.Connect(natsURL)
	if err != nil {
		return nil, err
	}

	return &natsPublisher{conn: conn}, nil
}

func (p *natsPublisher) PublishBookingCreated(event BookingCreatedEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.conn.Publish("booking.created", data)
}

func (p *natsPublisher) PublishBookingStatusChanged(event BookingStatusChangedEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.conn.Publish("booking.status_changed", data)
}
