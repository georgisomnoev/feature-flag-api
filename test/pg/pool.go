package pg

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	. "github.com/onsi/gomega"
)

const DBConnectionURL = "postgres://ffuser:ffpass@localhost:5432/featureflagsdb"

func MustInitDBPool(ctx context.Context) *pgxpool.Pool {
	poolCfg, err := pgxpool.ParseConfig(DBConnectionURL)
	Expect(err).NotTo(HaveOccurred())

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(pool.Ping(ctx)).To(Succeed())

	return pool
}
