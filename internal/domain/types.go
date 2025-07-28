package domain

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
)

type (
	UserID = uuid.UUID
	ServiceName        = string

	Subscription struct {
		Name      ServiceName        `json:"name" db:"service_name"`
		Cost      int                `json:"cost" db:"month_cost"`
		UserID    UserID `json:"id" db:"user_id"`
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
		Delete(context.Context, UserID, ServiceName) error
		ReadAllByUserID(context.Context, UserID) ([]Subscription, error)
		TotalSubscriptionsCost(context.Context, UserID, ServiceName, time.Time, *time.Time) (int, error)
		io.Closer
	}
)
