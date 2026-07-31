package service

import (
	"context"
	"passwordmanager/internal/dto"
)

type AuthServiceInterface interface {
	Register(ctx context.Context, login string, password string) error
	Login(ctx context.Context, login string, password string) (string, error)
}

type TokenManagerInterface interface {
	Generate(userId int64, login string) (string, error)
	Parse(token string) (*Claims, error)
}

type VaultServiceInterface interface {
	Create(ctx context.Context, userID int64, item VaultCreate) (int64, error)
	Get(ctx context.Context, userID, itemID int64) (dto.VaultResponseGet, error)
	List(ctx context.Context, userID int64) ([]dto.VaultResponseGet, error)
}
