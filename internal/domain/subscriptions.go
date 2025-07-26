package domain

import (
	"context"
	"errors"
	"time"
)

var _ SubscriptionInterface = (*SubscriptionService)(nil)

var (
	errServiseSubscription               = errors.New("service error")
	ErrServiceCreateSubscription         = errors.Join(errServiseSubscription, errors.New("create failed"))
	ErrServiceUpdateSubscription         = errors.Join(errServiseSubscription, errors.New("update failed"))
	ErrServiceDeleteSubscription         = errors.Join(errServiseSubscription, errors.New("delete failed"))
	ErrServiceReadAllByUserIDList        = errors.Join(errServiseSubscription, errors.New("read all by user id failed"))
	ErrServiceTotalSubscriptionsCostList = errors.Join(errServiseSubscription, errors.New("total cost failed"))
)

type SubscriptionService struct {
	provider         ConnectionProvider
	subscriptionRepo SubscriptionsRepository
}

func NewSubscriptionService(provider ConnectionProvider, subscriptionRepo SubscriptionsRepository) *SubscriptionService {
	return &SubscriptionService{
		provider:         provider,
		subscriptionRepo: subscriptionRepo,
	}
}

func (s *SubscriptionService) Create(ctx context.Context, subscription Subscription) error {
	err := s.provider.ExecuteTx(ctx, func(ctx context.Context, c Connection) error {
		return s.subscriptionRepo.Create(ctx, c, subscription)
	})
	if err != nil {
		return errors.Join(err, ErrServiceCreateSubscription)
	}
	return nil
}

func (s *SubscriptionService) Delete(ctx context.Context, subscriptionUserID SubscriptionUserID, subscriptionName ServiceName) error {
	err := s.provider.ExecuteTx(ctx, func(ctx context.Context, c Connection) error {
		return s.subscriptionRepo.Delete(ctx, c, subscriptionUserID, subscriptionName)
	})
	if err != nil {
		return errors.Join(err, ErrServiceCreateSubscription)
	}
	return nil
}

func (s *SubscriptionService) Update(ctx context.Context, subscription Subscription) error {
	err := s.provider.ExecuteTx(ctx, func(ctx context.Context, c Connection) error {
		return s.subscriptionRepo.Update(ctx, c, subscription)
	})
	if err != nil {
		return errors.Join(err, ErrServiceCreateSubscription)
	}
	return nil
}

func (s *SubscriptionService) ReadAllByUserID(ctx context.Context, subscriptionUserID SubscriptionUserID) ([]Subscription, error) {
	var subscriptions []Subscription
	var dbErr error
	err := s.provider.Execute(ctx, func(ctx context.Context, c Connection) error {
		subscriptions, dbErr = s.subscriptionRepo.ReadAllByUserID(ctx, c, subscriptionUserID)
		return dbErr
	})
	if err != nil {
		return subscriptions, errors.Join(err, ErrServiceCreateSubscription)
	}
	return subscriptions, nil
}

func (s *SubscriptionService) TotalSubscriptionsCost(ctx context.Context, subscriptionUserID SubscriptionUserID, subscriptionName ServiceName, start time.Time, end *time.Time) (int, error) {
	var totalCosts []int
	var dbErr error
	err := s.provider.Execute(ctx, func(ctx context.Context, c Connection) error {
		totalCosts, dbErr = s.subscriptionRepo.AllMatchingSubscriptionsForPeriod(ctx, c, subscriptionUserID, subscriptionName, start, end)
		return dbErr
	})
	totalCost := 0
	if err != nil {
		return totalCost, errors.Join(err, ErrServiceCreateSubscription)
	}
	for _, curCost := range totalCosts {
		totalCost += curCost
	}

	return totalCost, nil
}

func (s *SubscriptionService) Close() error {
	return s.provider.Close()
}
