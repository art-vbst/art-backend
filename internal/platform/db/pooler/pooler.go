package pooler

import (
	"context"
	"log"
	"time"

	"github.com/art-vbst/art-backend/internal/platform/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetDbConnectionPool(ctx context.Context, env *config.Config) *pgxpool.Pool {
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

	return pool
}
