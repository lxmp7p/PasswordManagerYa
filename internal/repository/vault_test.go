package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"passwordmanager/internal/model"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func TestVaultRepository_Create(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := NewVaultRepository(db)

	db.ExpectQuery("INSERT INTO vault_items").
		WithArgs(
			int64(1),
			model.ItemLogin,
			"test",
			[]byte("secret"),
		).
		WillReturnRows(
			pgxmock.NewRows([]string{"id"}).
				AddRow(int64(10)),
		)

	id, err := repo.Create(
		context.Background(),
		model.VaultItem{
			UserID:     1,
			Type:       model.ItemLogin,
			Title:      "test",
			SecretData: []byte("secret"),
		},
	)

	require.NoError(t, err)
	require.Equal(t, int64(10), id)

	require.NoError(t, db.ExpectationsWereMet())
}

func TestVaultRepository_GetByID(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := NewVaultRepository(db)

	now := time.Now()

	db.ExpectQuery("SELECT").
		WithArgs(
			int64(10),
			int64(1),
		).
		WillReturnRows(
			pgxmock.NewRows(
				[]string{
					"id",
					"user_id",
					"type",
					"title",
					"secret_data",
					"created_at",
					"updated_at",
				},
			).
				AddRow(
					10,
					1,
					model.ItemLogin,
					"github",
					[]byte("secret"),
					now,
					now,
				),
		)

	item, err := repo.GetByID(
		context.Background(),
		10,
		1,
	)

	require.NoError(t, err)
	require.Equal(t, int64(10), item.ID)
	require.Equal(t, "github", item.Title)

	require.NoError(t, db.ExpectationsWereMet())
}

func TestVaultRepository_CreateMetadata(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := NewVaultRepository(db)

	db.ExpectExec("INSERT INTO vault_metadata").
		WithArgs(
			int64(1),
			"login",
			"admin",
		).
		WillReturnResult(
			pgxmock.NewResult("INSERT", 1),
		)

	err = repo.CreateMetadata(
		context.Background(),
		1,
		map[string]string{
			"login": "admin",
		},
	)

	require.NoError(t, err)

	require.NoError(t, db.ExpectationsWereMet())
}

func TestVaultRepository_GetMetadata(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := NewVaultRepository(db)

	db.ExpectQuery("SELECT key, value").
		WithArgs(int64(1)).
		WillReturnRows(
			pgxmock.NewRows(
				[]string{"key", "value"},
			).
				AddRow("login", "admin"),
		)

	result, err := repo.GetMetadata(
		context.Background(),
		1,
	)

	require.NoError(t, err)

	require.Equal(
		t,
		map[string]string{
			"login": "admin",
		},
		result,
	)

	require.NoError(t, db.ExpectationsWereMet())
}

func TestVaultRepository_List(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := NewVaultRepository(db)

	now := time.Now()

	db.ExpectQuery("SELECT").
		WithArgs(int64(1)).
		WillReturnRows(
			pgxmock.NewRows(
				[]string{
					"id",
					"user_id",
					"type",
					"title",
					"secret_data",
					"created_at",
					"updated_at",
				},
			).
				AddRow(
					int64(10),
					int64(1),
					model.ItemLogin,
					"github",
					[]byte("secret"),
					now,
					now,
				),
		)

	result, err := repo.List(
		context.Background(),
		1,
	)

	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "github", result[0].Title)

	require.NoError(t, db.ExpectationsWereMet())
}

func TestVaultRepository_List_Error(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := NewVaultRepository(db)

	db.ExpectQuery("SELECT").
		WithArgs(int64(1)).
		WillReturnError(
			errors.New("db error"),
		)

	result, err := repo.List(
		context.Background(),
		1,
	)

	require.Error(t, err)
	require.Nil(t, result)

	require.NoError(t, db.ExpectationsWereMet())
}

func TestVaultRepository_GetByID_Error(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := NewVaultRepository(db)

	db.ExpectQuery("SELECT").
		WithArgs(
			int64(10),
			int64(1),
		).
		WillReturnError(
			errors.New("not found"),
		)

	item, err := repo.GetByID(
		context.Background(),
		10,
		1,
	)

	require.Error(t, err)
	require.Nil(t, item)

	require.NoError(t, db.ExpectationsWereMet())
}

func TestVaultRepository_CreateMetadata_Error(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := NewVaultRepository(db)

	db.ExpectExec("INSERT INTO vault_metadata").
		WithArgs(
			int64(1),
			"login",
			"admin",
		).
		WillReturnError(
			errors.New("insert error"),
		)

	err = repo.CreateMetadata(
		context.Background(),
		1,
		map[string]string{
			"login": "admin",
		},
	)

	require.Error(t, err)

	require.NoError(t, db.ExpectationsWereMet())
}

func TestVaultRepository_GetMetadata_Error(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := NewVaultRepository(db)

	db.ExpectQuery("SELECT key, value").
		WithArgs(int64(1)).
		WillReturnError(
			errors.New("db error"),
		)

	result, err := repo.GetMetadata(
		context.Background(),
		1,
	)

	require.Error(t, err)
	require.Nil(t, result)

	require.NoError(t, db.ExpectationsWereMet())
}
