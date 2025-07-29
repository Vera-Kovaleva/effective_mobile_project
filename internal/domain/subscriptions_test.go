package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"ef_project/internal/domain"
	"ef_project/internal/generated/mocks"
	"ef_project/internal/infra/pointer"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestServicePVZ_Create(t *testing.T) {
	t.Parallel()

	validSubscription := domain.Subscription{
		Name:      "service_name",
		Cost:      100,
		UserID:    uuid.New(),
		StartDate: time.Now().UTC().Truncate(24 * time.Hour),
		EndDate:   pointer.Ref(time.Now().UTC().Truncate(24 * time.Hour)),
	}

	tests := []struct {
		name         string
		subscribtion domain.Subscription
		prepareMocks func(*mocks.MockConnectionProvider, *mocks.MockSubscriptionsRepository)
		check        func(*testing.T, error)
	}{
		{
			name:         "Success",
			subscribtion: validSubscription,
			prepareMocks: func(provider *mocks.MockConnectionProvider, repo *mocks.MockSubscriptionsRepository) {
				provider.EXPECT().
					ExecuteTx(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, f func(context.Context, domain.Connection) error) error {
						return f(ctx, &mocks.MockConnection{})
					}).
					Once()

				repo.EXPECT().
					GetLatestSubscriptionDate(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(validSubscription.EndDate, nil).
					Once()
				repo.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
					Return(nil).Once()
			},
			check: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:         "DB create Error",
			subscribtion: validSubscription,
			prepareMocks: func(provider *mocks.MockConnectionProvider, repo *mocks.MockSubscriptionsRepository) {
				provider.EXPECT().
					ExecuteTx(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, f func(context.Context, domain.Connection) error) error {
						return f(ctx, &mocks.MockConnection{})
					}).
					Once()

				repo.EXPECT().
					GetLatestSubscriptionDate(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(validSubscription.EndDate, nil).
					Once()
				repo.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("some error")).Once()
			},
			check: func(t *testing.T, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "some error")
				require.Contains(t, err.Error(), "create failed")
			},
		},
		{
			name:         "DB get latest Error",
			subscribtion: validSubscription,
			prepareMocks: func(provider *mocks.MockConnectionProvider, repo *mocks.MockSubscriptionsRepository) {
				provider.EXPECT().
					ExecuteTx(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, f func(context.Context, domain.Connection) error) error {
						return f(ctx, &mocks.MockConnection{})
					}).
					Once()

				repo.EXPECT().
					GetLatestSubscriptionDate(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(validSubscription.EndDate, errors.New("some error")).
					Once()
			},
			check: func(t *testing.T, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "some error")
				require.Contains(t, err.Error(), "get latest date failed")
			},
		},
		{
			name:         "Subscription has not ended Error",
			subscribtion: validSubscription,
			prepareMocks: func(provider *mocks.MockConnectionProvider, repo *mocks.MockSubscriptionsRepository) {
				provider.EXPECT().
					ExecuteTx(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, f func(context.Context, domain.Connection) error) error {
						return f(ctx, &mocks.MockConnection{})
					}).
					Once()

				repo.EXPECT().
					GetLatestSubscriptionDate(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, nil).
					Once()
			},
			check: func(t *testing.T, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "some error")
				require.Contains(t, err.Error(), "previous subscription has not ende")
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			provider := mocks.NewMockConnectionProvider(t)

			repoSunbscriptions := mocks.NewMockSubscriptionsRepository(t)

			if test.prepareMocks != nil {
				test.prepareMocks(provider, repoSunbscriptions)
			}
			err := domain.NewSubscriptionService(provider, repoSunbscriptions).
				Create(t.Context(), validSubscription)

			test.check(t, err)
		})
	}
}
