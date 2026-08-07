package repository

import (
	"context"
	"passwordmanager/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type User struct {
	ID           int64
	Login        string
	PasswordHash string
}

type DBinterface interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type UserRepositoryInterface interface {
	Create(ctx context.Context, login string, passwordHash string) error
	FindByLogin(ctx context.Context, login string) (*User, error)
}

type VaultRepositoryInterface interface {
	Create(ctx context.Context, item model.VaultItem) (int64, error)
	CreateMetadata(ctx context.Context, itemID int64, metadata map[string]string) error

	GetByID(ctx context.Context, id, userID int64) (*model.VaultItem, error)
	GetMetadata(ctx context.Context, itemID int64) (map[string]string, error)

	List(ctx context.Context, userID int64) ([]model.VaultItem, error)
}
