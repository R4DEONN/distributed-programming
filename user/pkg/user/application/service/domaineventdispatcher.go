package service

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"

	"userservice/pkg/common/domain"
)

type domainEventDispatcher struct {
	ctx             context.Context
	eventDispatcher outbox.EventDispatcher[outbox.Event]
}

func (d *domainEventDispatcher) Dispatch(event domain.Event) error {
	return d.eventDispatcher.Dispatch(d.ctx, event)
}
