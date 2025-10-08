package pooler

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/talmage89/art-backend/internal/platform/config"
	"github.com/talmage89/art-backend/internal/platform/db/generated"
)

func GetDbConnectionPool(ctx context.Context, env *config.Config) (*generated.Queries, *pgxpool.Pool) {
	poolConfig, err := pgxpool.ParseConfig(env.DbUrl)
	if err != nil {
		log.Fatal(err)
	}

	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = time.Minute * 30
	poolConfig.MaxConnIdleTime = time.Minute * 5
	poolConfig.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatal(err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	queries := generated.New(pool)

	return queries, pool
}
