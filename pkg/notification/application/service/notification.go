package service

import (
	"context"

	"github.com/google/uuid"

	"notificationservice/pkg/notification/domain/model"
	"notificationservice/pkg/notification/domain/service"
)

type RepositoryProvider interface {
	NotificationRepository(ctx context.Context) model.NotificationRepository
}

type LockableUnitOfWork interface {
	Execute(ctx context.Context, lockNames []string, f func(provider RepositoryProvider) error) error
}

type NotificationService interface {
	SendNotification(ctx context.Context, userID uuid.UUID, title, message string) (uuid.UUID, error)
	ListNotifications(ctx context.Context, userID uuid.UUID) ([]model.Notification, error)
}

func NewNotificationService(luow LockableUnitOfWork) NotificationService {
	return &notificationServiceImpl{luow: luow}
}

type notificationServiceImpl struct {
	luow LockableUnitOfWork
}

func (s *notificationServiceImpl) SendNotification(ctx context.Context, userID uuid.UUID, title, message string) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.luow.Execute(ctx, nil, func(provider RepositoryProvider) error {
		domainSvc := service.NewNotificationService(provider.NotificationRepository(ctx))
		var err error
		id, err = domainSvc.SendNotification(userID, title, message)
		return err
	})
	return id, err
}

func (s *notificationServiceImpl) ListNotifications(ctx context.Context, userID uuid.UUID) ([]model.Notification, error) {
	var result []model.Notification
	err := s.luow.Execute(ctx, nil, func(provider RepositoryProvider) error {
		repo := provider.NotificationRepository(ctx)
		var err error
		result, err = repo.FindAll(model.FindSpec{UserID: &userID})
		return err
	})
	return result, err
}
