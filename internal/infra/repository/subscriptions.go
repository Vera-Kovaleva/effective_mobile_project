package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"ef_project/internal/domain"
)

var _ domain.SubscriptionsRepository = (*Subscription)(nil)

var (
	errSubscription              = errors.New("subscription repository error")
	ErrCreateSubscription        = errors.Join(errSubscription, errors.New("create failed"))
	ErrReadAllSubscriptions      = errors.Join(errSubscription, errors.New("read failed"))
	ErrUpdateSubscription        = errors.Join(errSubscription, errors.New("update failed"))
	ErrDeleteSubscription        = errors.Join(errSubscription, errors.New("delete failed"))
	ErrTotalCostSubscription     = errors.Join(errSubscription, errors.New("total cost failed"))
	ErrGetLatestSubscription     = errors.Join(errSubscription, errors.New("get latest subscription failed"))
	ErrGetLatestDateSubscription = errors.Join(errSubscription, errors.New("get latest date failed"))
)

type Subscription struct{}

func NewSubscription() *Subscription {
	return &Subscription{}
}

func (s *Subscription) Create(
	ctx context.Context,
	connection domain.Connection,
	subscription domain.Subscription,
) error {
	const query = `insert into subscriptions
	(service_name, month_cost, user_id, subs_start_date, subs_end_date)
	values
	($1, $2, $3, $4, $5)`

	if _, err := connection.ExecContext(ctx, query, subscription.Name, subscription.Cost, subscription.UserID, subscription.StartDate, subscription.EndDate); err != nil {
		return errors.Join(ErrCreateSubscription, err)
	}

	return nil
}

func (s *Subscription) ReadAllByUserID(
	ctx context.Context,
	connection domain.Connection,
	userID domain.UserID,
) ([]domain.Subscription, error) {
	const query = `select service_name, month_cost, user_id, subs_start_date, subs_end_date from subscriptions where user_id=$1`
	var allUserSubscriptions []domain.Subscription
	if err := connection.SelectContext(ctx, &allUserSubscriptions, query, userID); err != nil {
		return allUserSubscriptions, errors.Join(ErrReadAllSubscriptions, err)
	}
	return allUserSubscriptions, nil
}

func (s *Subscription) Update(
	ctx context.Context,
	connection domain.Connection,
	subscription domain.Subscription,
) error {
	const query = `update subscriptions set month_cost = $3, subs_end_date=$4  
	where service_name = $1 and user_id = $2 
	and subs_start_date = (select subs_start_date from subscriptions
	where user_id = $2 and service_name = $1 order by subs_start_date desc limit 1) `

	if _, err := connection.ExecContext(ctx, query, subscription.Name, subscription.UserID, subscription.Cost, subscription.EndDate); err != nil {
		return errors.Join(ErrUpdateSubscription, err)
	}

	return nil
}

func (s *Subscription) Delete(
	ctx context.Context,
	connection domain.Connection,
	subscriptionUserID domain.UserID,
	subscriptionName domain.ServiceName,
) error {
	const query = `delete from subscriptions where user_id = $1 and service_name = $2 
	and subs_start_date = (select subs_start_date from subscriptions where user_id = $1 and service_name = $2 order by subs_start_date desc limit 1)`
	rowsAffected, err := connection.ExecContext(ctx, query, subscriptionUserID, subscriptionName)
	if err != nil {
		return errors.Join(ErrDeleteSubscription, err)
	}
	if rowsAffected == 0 {
		return errors.Join(ErrDeleteSubscription, errors.New("no subscription found to delete"))
	}
	return nil
}

func (s *Subscription) GetLatest(
	ctx context.Context,
	connection domain.Connection,
	userID domain.UserID,
) (domain.Subscription, error) {
	var latestSubs domain.Subscription
	const query = `select service_name, month_cost, user_id, subs_start_date, subs_end_date from subscriptions
	where user_id = $1 order by subs_start_date desc limit 1`

	if err := connection.GetContext(ctx, &latestSubs, query, userID); err != nil {
		return latestSubs, errors.Join(ErrGetLatestSubscription, err)
	}
	return latestSubs, nil
}

func (s *Subscription) AllMatchingSubscriptionsForPeriod(
	ctx context.Context,
	connection domain.Connection,
	subscriptionUserID domain.UserID,
	subscriptionName domain.ServiceName,
	start time.Time,
	end *time.Time,
) ([]int, error) {
	const query = `select month_cost from subscriptions where user_id = $1 and service_name = $2 and subs_start_date < $3 and (subs_end_date is null or subs_end_date >= $4)`
	var matchesSubscriptions []int
	if err := connection.SelectContext(ctx, &matchesSubscriptions, query, subscriptionUserID, subscriptionName, end, start); err != nil {
		return matchesSubscriptions, errors.Join(ErrTotalCostSubscription, err)
	}
	return matchesSubscriptions, nil
}

func (s *Subscription) GetLatestSubscriptionDate(
	ctx context.Context,
	connection domain.Connection,
	userID domain.UserID,
	serviceName domain.ServiceName,
) (*time.Time, error) {
	const query = `select (subs_end_date) from subscriptions
	where user_id = $1 and service_name = $2 order by subs_start_date desc limit 1`
	var latestDate *time.Time
	if err := connection.GetContext(ctx, &latestDate, query, userID, serviceName); err != nil &&
		!errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Join(ErrGetLatestDateSubscription, err)
	}
	return latestDate, nil
}
