package mysql

import (
	"context"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"

	"notificationservice/pkg/notification/application/service"
)

func NewLockableUnitOfWork(uow mysql.LockableUnitOfWorkWithRepositoryProvider[service.RepositoryProvider]) service.LockableUnitOfWork {
	return &lockableUnitOfWork{
		uow: uow,
	}
}

type lockableUnitOfWork struct {
	uow mysql.LockableUnitOfWorkWithRepositoryProvider[service.RepositoryProvider]
}

func (l *lockableUnitOfWork) Execute(ctx context.Context, lockNames []string, f func(provider service.RepositoryProvider) error) error {
	if len(lockNames) == 0 {
		return l.uow.ExecuteWithRepositoryProvider(ctx, "", time.Minute, f)
	}
	if len(lockNames) == 1 {
		return l.uow.ExecuteWithRepositoryProvider(ctx, lockNames[0], time.Minute, f)
	}
	ln := lockNames[0]
	lns := lockNames[1:]
	return l.uow.ExecuteWithRepositoryProvider(ctx, ln, time.Minute, func(_ service.RepositoryProvider) error {
		return l.Execute(ctx, lns, f)
	})
}
