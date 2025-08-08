package component

import (
	"context"
	"fmt"
	"time"
)

const timeout = 2 * time.Second

type DBConnector interface {
	Ping(ctx context.Context) error
}

type DBComponent struct {
	db DBConnector
}

func NewDBComponent(db DBConnector) *DBComponent {
	return &DBComponent{db: db}
}

func (d *DBComponent) Name() string {
	return "database"
}

func (d *DBComponent) Check() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := d.db.Ping(ctx); err != nil {
		return fmt.Errorf("database unreachable: %w", err)
	}
	return nil
}
