package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
)

type Notification struct {
	NotificationID uuid.UUID
	UserID         uuid.UUID
	Title          string
	Message        string
	CreatedAt      time.Time
}

type FindSpec struct {
	UserID *uuid.UUID
}

type NotificationRepository interface {
	NextID() (uuid.UUID, error)
	Store(notification Notification) error
	FindAll(spec FindSpec) ([]Notification, error)
}
