package usecase

import (
	"context"
	"database/sql"
	"time"
)

type HealthUC struct {
	db *sql.DB
}
type HealthUsecase interface {
	Check() error
}

func NewHealthUC(db *sql.DB) HealthUsecase {
	return &HealthUC{db}
}

func (u *HealthUC) Check() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := u.db.PingContext(ctx); err != nil {
		return ErrInternal
	}

	return nil
}
