package service

import (
	"time"

	"github.com/google/uuid"

	"notificationservice/pkg/notification/domain/model"
)

type NotificationService interface {
	SendNotification(userID uuid.UUID, title, message string) (uuid.UUID, error)
}

func NewNotificationService(repo model.NotificationRepository) NotificationService {
	return &notificationService{repo: repo}
}

type notificationService struct {
	repo model.NotificationRepository
}

func (s *notificationService) SendNotification(userID uuid.UUID, title, message string) (uuid.UUID, error) {
	id, err := s.repo.NextID()
	if err != nil {
		return uuid.Nil, err
	}

	notification := model.Notification{
		NotificationID: id,
		UserID:         userID,
		Title:          title,
		Message:        message,
		CreatedAt:      time.Now(),
	}

	err = s.repo.Store(notification)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}
