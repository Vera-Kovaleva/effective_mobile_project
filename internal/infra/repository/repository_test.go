package repository_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"ef_project/internal/domain"
	"ef_project/internal/infra/database"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func rollback(t *testing.T, test func(context.Context, domain.Connection)) {
	provider := cleanTablesAndCreateProvider(t)
	defer provider.Close()

	err := provider.ExecuteTx(
		t.Context(),
		func(ctx context.Context, connection domain.Connection) error {
			test(ctx, connection)

			return nil
		},
	)
	require.NoError(t, err)
}

func cleanTablesAndCreateProvider(t *testing.T) domain.ConnectionProvider {
	_ = godotenv.Load("../../../.env")

	tablesToClean := []string{"subscriptions"}

	pool, err := pgxpool.New(context.Background(), os.Getenv("DB_CONNECTION"))
	require.NoError(t, err)

	provider := database.NewPostgresProvider(pool)

	err = provider.Execute(
		t.Context(),
		func(ctx context.Context, connection domain.Connection) error {
			for _, table := range tablesToClean {
				_, err = connection.ExecContext(ctx, fmt.Sprintf("delete from %s cascade", table))
				require.NoError(t, err)
			}

			return nil
		},
	)
	require.NoError(t, err)

	return provider
}
