package impl

import (
	"context"
	"fmt"
	"os"
	"portal-system/internal/infrastructure/storage"
	"testing"
	"time"

	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"

	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	os.Exit(runTestMain(m))
}

func runTestMain(m *testing.M) int {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := tcpostgres.Run(
		ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("portal_test"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start postgres test container: %v\n", err)
		return 1
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "postgres connection string: %v\n", err)
		_ = container.Terminate(context.Background())
		return 1
	}

	testDB, err = gorm.Open(pgdriver.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "open gorm db: %v\n", err)
		_ = container.Terminate(context.Background())
		return 1
	}

	if err := storage.AutoMigrate(testDB); err != nil {
		fmt.Fprintf(os.Stderr, "auto migrate test db: %v\n", err)
		_ = container.Terminate(context.Background())
		return 1
	}

	code := m.Run()

	if sqlDB, err := testDB.DB(); err == nil {
		_ = sqlDB.Close()
	}
	_ = container.Terminate(context.Background())

	return code
}

func newTestTx(t *testing.T) (context.Context, *gorm.DB) {
	t.Helper()

	tx := testDB.Begin()
	if tx.Error != nil {
		t.Fatalf("begin tx: %v", tx.Error)
	}

	t.Cleanup(func() {
		_ = tx.Rollback().Error
	})

	return context.WithValue(context.Background(), txKey{}, tx), tx
}
