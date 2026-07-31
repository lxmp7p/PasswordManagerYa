package service

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"passwordmanager/internal/model"
	"passwordmanager/internal/repository"
)

type VaultCreate struct {
	Type     model.ItemType
	Title    string
	Data     []byte
	Metadata map[string]string
}

type VaultService struct {
	logger    *slog.Logger
	vaultRepo repository.VaultRepositoryInterface
}

func NewVaultService(
	logger *slog.Logger,
	vaultRepo repository.VaultRepositoryInterface,
) VaultServiceInterface {
	return &VaultService{
		logger:    logger,
		vaultRepo: vaultRepo,
	}
}

type LoginPasswordData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type BankCardData struct {
	Number string `json:"number"`
	Holder string `json:"holder"`
	Month  int    `json:"month"`
	Year   int    `json:"year"`
	CVV    string `json:"cvv"`
}

type TextData struct {
	Text string `json:"text"`
}

func (vi *VaultService) Create(ctx context.Context, userID int64, item VaultCreate) error {
	switch item.Type {
	case model.ItemLogin:
		var data LoginPasswordData
		err := json.Unmarshal(item.Data, &data)
		if err != nil {
			return err
		}

		if err := data.validate(); err != nil {
			return err
		}

	case model.ItemBankCard:
		var data BankCardData
		err := json.Unmarshal(item.Data, &data)
		if err != nil {
			return err
		}

		if err := data.validate(); err != nil {
			return err
		}

	case model.ItemText:
		var data TextData
		err := json.Unmarshal(item.Data, &data)
		if err != nil {
			return err
		}

		if err := data.validate(); err != nil {
			return err
		}
	}

	vaultItem := model.VaultItem{
		UserID:     userID,
		Type:       item.Type,
		Title:      item.Title,
		SecretData: item.Data,
	}

	vi.vaultRepo.Create(ctx, vaultItem)

	return nil
}

func (lp *LoginPasswordData) validate() error {
	if lp.Login == "" || lp.Password == "" {
		return errors.New("login and password are required")
	}
	return nil
}

func (bc *BankCardData) validate() error {
	if bc.Number == "" || bc.Holder == "" || bc.Month == 0 || bc.Year == 0 || bc.CVV == "" {
		return errors.New("all fields are requered")
	}
	return nil
}

func (td *TextData) validate() error {
	if td.Text == "" {
		return errors.New("text are required")
	}
	return nil
}
