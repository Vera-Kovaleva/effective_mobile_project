package domain

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"ef_project/internal/infra/log"
)

var _ SubscriptionInterface = (*SubscriptionService)(nil)

var (
	errServiseSubscription       = errors.New("service error")
	ErrServiceCreateSubscription = errors.Join(
		errServiseSubscription,
		errors.New("create failed"),
	)
	ErrGetLatestSubscription = errors.Join(
		errServiseSubscription,
		errors.New("get latest subscription failed"),
	)
	ErrServiceUpdateSubscription = errors.Join(
		errServiseSubscription,
		errors.New("update failed"),
	)
	ErrServiceDeleteSubscription = errors.Join(
		errServiseSubscription,
		errors.New("delete failed"),
	)
	ErrServiceReadAllByUserIDList = errors.Join(
		errServiseSubscription,
		errors.New("read all by user id failed"),
	)
	ErrServiceTotalSubscriptionsCostList = errors.Join(
		errServiseSubscription,
		errors.New("total cost failed"),
	)
)

type SubscriptionService struct {
	provider         ConnectionProvider
	subscriptionRepo SubscriptionsRepository
}

func NewSubscriptionService(
	provider ConnectionProvider,
	subscriptionRepo SubscriptionsRepository,
) *SubscriptionService {
	return &SubscriptionService{
		provider:         provider,
		subscriptionRepo: subscriptionRepo,
	}
}

func (s *SubscriptionService) GetLatest(
	ctx context.Context,
	subscriptionUserID UserID,
) (Subscription, error) {
	slog.DebugContext(ctx, "Service: getting latest subscribtion.", log.RequestID(ctx))
	var subscription Subscription
	err := s.provider.Execute(ctx, func(ctx context.Context, c Connection) error {
		var dbError error
		subscription, dbError = s.subscriptionRepo.GetLatest(ctx, c, subscriptionUserID)
		return dbError
	})
	if err != nil {
		return subscription, errors.Join(ErrServiceDeleteSubscription, err)
	}
	return subscription, nil
}

func (s *SubscriptionService) Create(ctx context.Context, subscription Subscription) error {
	slog.DebugContext(ctx, "Service: creating subscription.", log.RequestID(ctx))
	err := s.provider.ExecuteTx(ctx, func(ctx context.Context, c Connection) error {
		latestEndDate, err := s.subscriptionRepo.GetLatestSubscriptionDate(
			ctx,
			c,
			subscription.UserID,
			subscription.Name,
		)
		if err != nil {
			return errors.Join(ErrGetLatestSubscription, err)
		}
		if latestEndDate != nil && (latestEndDate.After(subscription.StartDate)) {
			return errors.Join(
				ErrServiceCreateSubscription,
				errors.New("previous subscription has not ended"))
		}

		return s.subscriptionRepo.Create(ctx, c, subscription)
	})
	if err != nil {
		return errors.Join(ErrServiceCreateSubscription, err)
	}
	return nil
}

func (s *SubscriptionService) Delete(
	ctx context.Context,
	subscriptionUserID UserID,
	subscriptionName ServiceName,
) error {
	slog.DebugContext(ctx, "Service: deleting subscribtion.", log.RequestID(ctx))
	err := s.provider.Execute(ctx, func(ctx context.Context, c Connection) error {
		return s.subscriptionRepo.Delete(ctx, c, subscriptionUserID, subscriptionName)
	})
	if err != nil {
		return errors.Join(ErrServiceDeleteSubscription, err)
	}
	return nil
}

func (s *SubscriptionService) Update(ctx context.Context, subscription Subscription) error {
	slog.DebugContext(ctx, "Service: updating subscribtion.", log.RequestID(ctx))
	err := s.provider.Execute(ctx, func(ctx context.Context, c Connection) error {
		return s.subscriptionRepo.Update(ctx, c, subscription)
	})
	if err != nil {
		return errors.Join(ErrServiceUpdateSubscription, err)
	}
	return nil
}

func (s *SubscriptionService) ReadAllByUserID(
	ctx context.Context,
	subscriptionUserID UserID,
) ([]Subscription, error) {
	slog.DebugContext(ctx, "Service: reading subscribtions by user ID.", log.RequestID(ctx))
	var subscriptions []Subscription
	err := s.provider.Execute(ctx, func(ctx context.Context, c Connection) error {
		var dbErr error
		subscriptions, dbErr = s.subscriptionRepo.ReadAllByUserID(ctx, c, subscriptionUserID)
		return dbErr
	})
	if err != nil {
		return subscriptions, errors.Join(ErrServiceReadAllByUserIDList, err)
	}
	return subscriptions, nil
}

func (s *SubscriptionService) TotalSubscriptionsCost(
	ctx context.Context,
	subscriptionUserID UserID,
	subscriptionName ServiceName,
	start time.Time,
	end *time.Time,
) (int, error) {
	var totalCosts []int
	err := s.provider.Execute(ctx, func(ctx context.Context, c Connection) error {
		slog.DebugContext(ctx, "Service: calculating total cost.", log.RequestID(ctx))
		var dbErr error
		totalCosts, dbErr = s.subscriptionRepo.AllMatchingSubscriptionsForPeriod(
			ctx,
			c,
			subscriptionUserID,
			subscriptionName,
			start,
			end,
		)
		return dbErr
	})
	totalCost := 0
	if err != nil {
		return totalCost, errors.Join(ErrServiceTotalSubscriptionsCostList, err)
	}
	for _, curCost := range totalCosts {
		totalCost += curCost
	}
	return totalCost, nil
}
