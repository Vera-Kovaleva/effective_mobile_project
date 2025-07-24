package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type (
	SubscriptionID   = uuid.UUID
	SubscriptionName = string

	Subscription struct {
		Name      SubscriptionName `json:"name"`
		Cost      int              `json:"cost"`
		ID        SubscriptionID   `json:"id"`
		StartDate time.Time        `json:"start_date"`
		EndDate   *time.Time       `json:"end_date,omitempty"`
	}

	Connection interface {
		GetContext(context.Context, any, string, ...any) error
		SelectContext(context.Context, any, string, ...any) error
		ExecContext(context.Context, string, ...any) (int64, error)
	}
)
