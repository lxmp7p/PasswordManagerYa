package repository

import (
	"context"
	"passwordmanager/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type vaultRepository struct {
	db *pgxpool.Pool
}

func NewVaultRepository(db *pgxpool.Pool) VaultRepositoryInterface {
	return &vaultRepository{
		db: db,
	}
}

var _ VaultRepositoryInterface = (*vaultRepository)(nil)

func (r *vaultRepository) Create(ctx context.Context, item model.VaultItem) error {

	_, err := r.db.Exec(
		ctx,
		`INSERT INTO users(login, password_hash)
		 VALUES($1, $2)`,
	)

	return err
}
