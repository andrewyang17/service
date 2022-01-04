package commands

import (
	"context"
	"errors"
	"fmt"
	"github.com/andrewyang17/service/business/data/dbschema"
	"time"

	"github.com/andrewyang17/service/business/sys/database"
)

var ErrHelp = errors.New("provided help")

func Migrate(cfg database.Config) error {
	db, err := database.Open(cfg)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := dbschema.Migrate(ctx, db); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}

	fmt.Println("migrations complete")
	return nil
}
