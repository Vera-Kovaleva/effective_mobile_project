package domain

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
)

type (
	SubscriptionUserID = uuid.UUID
	ServiceName        = string

	Subscription struct {
		Name      ServiceName        `json:"name" db:"service_name"`
		Cost      int                `json:"cost" db:"month_cost"`
		UserID    SubscriptionUserID `json:"id" db:"user_id"`
		StartDate time.Time          `json:"start_date" db:"subs_start_date"`
		EndDate   *time.Time         `json:"end_date,omitempty" db:"subs_end_date"`
	}

	Connection interface {
		GetContext(context.Context, any, string, ...any) error
		SelectContext(context.Context, any, string, ...any) error
		ExecContext(context.Context, string, ...any) (int64, error)
	}
	ConnectionProvider interface {
		Execute(context.Context, func(context.Context, Connection) error) error
		ExecuteTx(context.Context, func(context.Context, Connection) error) error
		io.Closer
	}

	SubscriptionInterface interface {
		Create(context.Context, Subscription) error
		Update(context.Context, Subscription) error
		Delete(context.Context, SubscriptionUserID, ServiceName) error
		ReadAllByUserID(context.Context, SubscriptionUserID) ([]Subscription, error)
		TotalSubscriptionsCost(context.Context, SubscriptionUserID, ServiceName, time.Time, *time.Time) (int, error)
		io.Closer
	}
)
