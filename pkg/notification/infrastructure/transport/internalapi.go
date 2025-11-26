package transport

import (
	"context"

	"github.com/google/uuid"

	"notificationservice/api/server/notificationinternal"
	"notificationservice/pkg/notification/application/service"
)

func NewNotificationInternalAPI(svc service.NotificationService) notificationinternal.NotificationInternalServiceServer {
	return &notificationInternalAPI{svc: svc}
}

type notificationInternalAPI struct {
	svc service.NotificationService
	notificationinternal.UnimplementedNotificationInternalServiceServer
}

func (a *notificationInternalAPI) SendNotification(ctx context.Context, req *notificationinternal.SendNotificationRequest) (*notificationinternal.SendNotificationResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, err
	}
	id, err := a.svc.SendNotification(ctx, userID, req.Title, req.Message)
	if err != nil {
		return nil, err
	}
	return &notificationinternal.SendNotificationResponse{NotificationID: id.String()}, nil
}

func (a *notificationInternalAPI) ListNotifications(ctx context.Context, req *notificationinternal.ListNotificationsRequest) (*notificationinternal.ListNotificationsResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, err
	}
	notifications, err := a.svc.ListNotifications(ctx, userID)
	if err != nil {
		return nil, err
	}

	items := make([]*notificationinternal.NotificationItem, 0, len(notifications))
	for _, n := range notifications {
		items = append(items, &notificationinternal.NotificationItem{
			NotificationID: n.NotificationID.String(),
			UserID:         n.UserID.String(),
			Title:          n.Title,
			Message:        n.Message,
			CreatedAt:      n.CreatedAt.Unix(),
		})
	}

	return &notificationinternal.ListNotificationsResponse{Notifications: items}, nil
}
