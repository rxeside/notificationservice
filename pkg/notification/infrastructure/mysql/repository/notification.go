package repository

import (
	"context"
	"strings"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"notificationservice/pkg/notification/domain/model"
)

func NewNotificationRepository(ctx context.Context, client mysql.ClientContext) model.NotificationRepository {
	return &notificationRepository{
		ctx:    ctx,
		client: client,
	}
}

type notificationRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

func (r *notificationRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (r *notificationRepository) Store(n model.Notification) error {
	_, err := r.client.ExecContext(r.ctx,
		`INSERT INTO notifications (notification_id, user_id, title, message, created_at) VALUES (?, ?, ?, ?, ?)`,
		n.NotificationID,
		n.UserID,
		n.Title,
		n.Message,
		n.CreatedAt,
	)
	return errors.WithStack(err)
}

func (r *notificationRepository) FindAll(spec model.FindSpec) ([]model.Notification, error) {
	var dtos []struct {
		NotificationID uuid.UUID `db:"notification_id"`
		UserID         uuid.UUID `db:"user_id"`
		Title          string    `db:"title"`
		Message        string    `db:"message"`
		CreatedAt      time.Time `db:"created_at"`
	}
	query, args := r.buildSpecArgs(spec)

	err := r.client.SelectContext(
		r.ctx,
		&dtos,
		`SELECT notification_id, user_id, title, message, created_at FROM notifications WHERE `+query+` ORDER BY created_at DESC`,
		args...,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result := make([]model.Notification, 0, len(dtos))
	for _, dto := range dtos {
		result = append(result, model.Notification{
			NotificationID: dto.NotificationID,
			UserID:         dto.UserID,
			Title:          dto.Title,
			Message:        dto.Message,
			CreatedAt:      dto.CreatedAt,
		})
	}
	return result, nil
}

func (r *notificationRepository) buildSpecArgs(spec model.FindSpec) (query string, args []interface{}) {
	var parts []string
	if spec.UserID != nil {
		parts = append(parts, "user_id = ?")
		args = append(args, *spec.UserID)
	}
	if len(parts) == 0 {
		return "1=1", nil
	}
	return strings.Join(parts, " AND "), args
}
