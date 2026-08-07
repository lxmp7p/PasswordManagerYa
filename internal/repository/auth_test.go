package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_Create(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	db.ExpectExec("INSERT INTO users").
		WithArgs(
			"admin",
			"hash",
		).
		WillReturnResult(
			pgxmock.NewResult("INSERT", 1),
		)

	err = repo.Create(
		context.Background(),
		"admin",
		"hash",
	)

	require.NoError(t, err)

	require.NoError(t, db.ExpectationsWereMet())
}
func TestUserRepository_Create_Error(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	db.ExpectExec("INSERT INTO users").
		WithArgs(
			"admin",
			"hash",
		).
		WillReturnError(
			errors.New("db error"),
		)

	err = repo.Create(
		context.Background(),
		"admin",
		"hash",
	)

	require.Error(t, err)

	require.NoError(t, db.ExpectationsWereMet())
}
func TestUserRepository_FindByLogin(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	db.ExpectQuery("SELECT id, login, password_hash").
		WithArgs("admin").
		WillReturnRows(
			pgxmock.NewRows(
				[]string{
					"id",
					"login",
					"password_hash",
				},
			).
				AddRow(
					int64(1),
					"admin",
					"hash",
				),
		)

	user, err := repo.FindByLogin(
		context.Background(),
		"admin",
	)

	require.NoError(t, err)

	require.Equal(t, int64(1), user.ID)
	require.Equal(t, "admin", user.Login)
	require.Equal(t, "hash", user.PasswordHash)

	require.NoError(t, db.ExpectationsWereMet())
}
func TestUserRepository_FindByLogin_Error(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	db.ExpectQuery("SELECT id, login, password_hash").
		WithArgs("admin").
		WillReturnError(
			errors.New("not found"),
		)

	user, err := repo.FindByLogin(
		context.Background(),
		"admin",
	)

	require.Error(t, err)
	require.Nil(t, user)

	require.NoError(t, db.ExpectationsWereMet())
}
