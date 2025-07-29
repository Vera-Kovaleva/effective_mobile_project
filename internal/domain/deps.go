package domain

import (
	"context"
	"time"
)

type SubscriptionsRepository interface {
	Create(context.Context, Connection, Subscription) error
	Update(context.Context, Connection, Subscription) error
	Delete(context.Context, Connection, UserID, ServiceName) error
	ReadAllByUserID(context.Context, Connection, UserID) ([]Subscription, error)
	GetLatest(context.Context, Connection, UserID) (Subscription, error)
	AllMatchingSubscriptionsForPeriod(
		context.Context,
		Connection,
		UserID,
		ServiceName,
		time.Time,
		*time.Time,
	) ([]int, error)
	GetLatestSubscriptionDate(context.Context, Connection, UserID, ServiceName) (*time.Time, error)
}
