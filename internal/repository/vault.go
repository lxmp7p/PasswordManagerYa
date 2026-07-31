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

func (r *vaultRepository) Create(
	ctx context.Context,
	item model.VaultItem,
) error {

	_, err := r.db.Exec(
		ctx,
		`
		INSERT INTO vault_items(
			user_id,
			type,
			title,
			secret_data
		)
		VALUES($1, $2, $3, $4)
		`,
		item.UserID,
		item.Type,
		item.Title,
		item.SecretData,
	)

	return err
}
