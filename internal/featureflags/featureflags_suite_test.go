package featureflags_test

import (
	"context"
	"testing"

	testdb "github.com/georgisomnoev/feature-flag-api/test/pg"
	"github.com/jackc/pgx/v5/pgxpool"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	ctx  context.Context
	pool *pgxpool.Pool
)

func TestFeatureFlags(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Feature Flags Integration Suite")
}

var _ = BeforeSuite(func() {
	ctx = context.Background()
	pool = testdb.MustInitDBPool(ctx)
})

var _ = AfterSuite(func() {
	pool.Close()
})
