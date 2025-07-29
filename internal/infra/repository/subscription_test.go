package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"ef_project/internal/domain"
	"ef_project/internal/infra/pointer"
	"ef_project/internal/infra/repository"

	"ef_project/internal/generated/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSubscriptionIntegration(t *testing.T) {
	ctx := context.Background()
	repoSubscription := repository.NewSubscription()
	provider := cleanTablesAndCreateProvider(ctx, t)
	defer func() { _ = provider.Close() }()

	err := provider.ExecuteTx(ctx, func(ctx context.Context, connection domain.Connection) error {
		userID1 := domain.UserID(uuid.New())
		userID2 := domain.UserID(uuid.New())

		serviseName1 := domain.ServiceName("servise name 1")
		serviseName2 := domain.ServiceName("servise name 2")

		subscription1User1 := fixtureCreateSubscription(t, ctx, connection, userID1, serviseName1)
		_ = fixtureCreateSubscription(t, ctx, connection, userID1, serviseName2)

		subscription1User2 := fixtureCreateSubscription(t, ctx, connection, userID2, serviseName1)

		subscriptionsFromDBUser1, err := repoSubscription.ReadAllByUserID(ctx, connection, userID1)
		require.NoError(t, err)

		subscriptionsFromDBUser2, err := repoSubscription.ReadAllByUserID(ctx, connection, userID2)

		require.NoError(t, err)
		require.Len(t, subscriptionsFromDBUser1, 2)
		require.Len(t, subscriptionsFromDBUser2, 1)

		require.Equal(t, subscription1User1, subscriptionsFromDBUser1[0])

		err = repoSubscription.Delete(ctx, connection, userID1, serviseName1)
		require.NoError(t, err)

		subscriptionsFromDBUser1, err = repoSubscription.ReadAllByUserID(ctx, connection, userID1)
		require.NoError(t, err)
		require.Len(t, subscriptionsFromDBUser1, 1)

		now := time.Now()
		newEndDate := now.AddDate(0, 1, 0).UTC().Truncate(24 * time.Hour)
		subscription1User2.EndDate = pointer.Ref(newEndDate)
		err = repoSubscription.Update(ctx, connection, subscription1User2)
		require.NoError(t, err)

		subscriptionsFromDBUser2, err = repoSubscription.ReadAllByUserID(
			ctx,
			connection,
			subscription1User2.UserID,
		)
		require.NoError(t, err)
		require.Equal(t, pointer.Ref(newEndDate), subscriptionsFromDBUser2[0].EndDate)

		subscription2User2 := fixtureCreateSubscription(t, ctx, connection, userID2, serviseName2)
		subscriptionsCosts, err := repoSubscription.AllMatchingSubscriptionsForPeriod(
			ctx,
			connection,
			userID2,
			serviseName2,
			now,
			pointer.Ref(newEndDate),
		)
		require.NoError(t, err)
		require.Equal(t, subscription2User2.Cost, subscriptionsCosts[0])

		return nil
	})

	require.NoError(t, err)

}

func TestListsUnit(t *testing.T) {
	validSubscription := domain.Subscription{
		Name:      "service_name",
		Cost:      100,
		UserID:    uuid.New(),
		StartDate: time.Now().UTC().Truncate(24 * time.Hour),
		EndDate:   pointer.Ref(time.Now().UTC().Truncate(24 * time.Hour)),
	}
	ctx := context.Background()

	tests := []struct {
		name  string
		check func(*testing.T, *repository.Subscription, *mocks.MockConnection)
	}{
		{
			name: "Create Subscription Error",
			check: func(t *testing.T, repo *repository.Subscription, connection *mocks.MockConnection) {
				connection.EXPECT().
					ExecContext(mock.Anything, mock.Anything, mock.Anything).
					Return(0, errors.New("some error")).
					Once()

				err := repo.Create(ctx, connection, validSubscription)

				require.ErrorIs(t, err, repository.ErrCreateSubscription)
				require.ErrorContains(t, err, "some error")
			},
		},
		{
			name: "Delete Subscription Error",
			check: func(t *testing.T, repo *repository.Subscription, connection *mocks.MockConnection) {
				connection.EXPECT().
					ExecContext(mock.Anything, mock.Anything, mock.Anything).
					Return(0, errors.New("some error")).
					Once()

				err := repo.Delete(ctx, connection, validSubscription.UserID, validSubscription.Name)

				require.ErrorIs(t, err, repository.ErrDeleteSubscription)
				require.ErrorContains(t, err, "some error")
			},
		},
		{
			name: "Update Subscription Error",
			check: func(t *testing.T, repo *repository.Subscription, connection *mocks.MockConnection) {
				connection.EXPECT().
					ExecContext(mock.Anything, mock.Anything, mock.Anything).
					Return(0, errors.New("some error")).
					Once()

				err := repo.Update(ctx, connection, validSubscription)

				require.ErrorIs(t, err, repository.ErrUpdateSubscription)
				require.ErrorContains(t, err, "some error")
			},
		},
		{
			name: "Read All Subscriptions Error",
			check: func(t *testing.T, repo *repository.Subscription, connection *mocks.MockConnection) {
				connection.EXPECT().
					SelectContext(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("some error")).
					Once()

				_, err := repo.ReadAllByUserID(ctx, connection, validSubscription.UserID)

				require.ErrorIs(t, err, repository.ErrReadSubscription)
				require.ErrorContains(t, err, "some error")
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.check(t, repository.NewSubscription(), mocks.NewMockConnection(t))
		})
	}
}

func fixtureCreateSubscription(
	t *testing.T,
	ctx context.Context,
	connection domain.Connection,
	userID domain.UserID,
	name domain.ServiceName,
) domain.Subscription {
	subscription := domain.Subscription{
		UserID:    userID,
		Cost:      1,
		Name:      name,
		StartDate: time.Now().UTC().Truncate(24 * time.Hour),
	}
	require.NoError(t, repository.NewSubscription().Create(ctx, connection, subscription))

	return subscription
}
