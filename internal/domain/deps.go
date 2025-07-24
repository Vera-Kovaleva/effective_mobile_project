package domain

import "context"

type SubscriptionsRepository interface {
	Create(context.Context, Connection, Subscription) error
	Update(context.Context, Connection, Subscription) error
	Delete(context.Context, Connection, SubscriptionID, SubscriptionName) error
	ReadAll(context.Context, Connection, SubscriptionID) ([]Subscription, error)
}
