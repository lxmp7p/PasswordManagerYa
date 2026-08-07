package repository

import (
	"context"
	"passwordmanager/internal/model"
)

type vaultRepository struct {
	db DBinterface
}

func NewVaultRepository(db DBinterface) VaultRepositoryInterface {
	return &vaultRepository{
		db: db,
	}
}

var _ VaultRepositoryInterface = (*vaultRepository)(nil)

func (r *vaultRepository) Create(ctx context.Context, item model.VaultItem) (int64, error) {
	var id int64
	err := r.db.QueryRow(
		ctx,
		`
		INSERT INTO vault_items(
			user_id,
			type,
			title,
			secret_data
		)
		VALUES($1, $2, $3, $4)
		RETURNING id;
		`,
		item.UserID,
		item.Type,
		item.Title,
		item.SecretData,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *vaultRepository) GetByID(ctx context.Context, id, userID int64) (*model.VaultItem, error) {
	item := &model.VaultItem{}
	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			user_id,
			type,
			title,
			secret_data,
			created_at,
			updated_at
		FROM vault_items
		WHERE id = $1
		  AND user_id = $2
		`,
		id,
		userID,
	).Scan(
		&item.ID,
		&item.UserID,
		&item.Type,
		&item.Title,
		&item.SecretData,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return item, nil
}

func (r *vaultRepository) List(ctx context.Context, userID int64) ([]model.VaultItem, error) {
	rows, err := r.db.Query(ctx,
		`
		SELECT
			id,
			user_id,
			type,
			title,
			secret_data,
			created_at,
			updated_at
		FROM vault_items
		WHERE user_id=$1
		ORDER BY updated_at DESC;`,
		userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.VaultItem

	for rows.Next() {
		var item model.VaultItem

		err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.Type,
			&item.Title,
			&item.SecretData,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *vaultRepository) CreateMetadata(ctx context.Context, itemID int64, metadata map[string]string) error {
	for key, value := range metadata {
		_, err := r.db.Exec(
			ctx,
			`
			INSERT INTO vault_metadata(
				item_id,
				key,
				value
			)
			VALUES($1, $2, $3)
			`,
			itemID,
			key,
			value,
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func (r *vaultRepository) GetMetadata(ctx context.Context, itemID int64) (map[string]string, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT key, value
		FROM vault_metadata
		WHERE item_id=$1
		`,
		itemID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metadata := make(map[string]string)

	for rows.Next() {
		var key string
		var value string

		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}

		metadata[key] = value
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return metadata, nil
}
