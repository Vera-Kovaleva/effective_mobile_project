package domain

import (
	"context"
	"time"
)

type SubscriptionsRepository interface {
	Create(context.Context, Connection, Subscription) error
	Update(context.Context, Connection, Subscription) error
	Delete(context.Context, Connection, SubscriptionUserID, ServiceName) error
	ReadAllByUserID(context.Context, Connection, SubscriptionUserID) ([]Subscription, error)
	AllMatchingSubscriptionsForPeriod(context.Context, Connection, SubscriptionUserID, ServiceName, time.Time, *time.Time) ([]int, error)
}
