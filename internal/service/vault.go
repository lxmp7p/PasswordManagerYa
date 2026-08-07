package service

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"passwordmanager/internal/dto"
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

type FileData struct {
	Name string `json:"name"`
	Data []byte `json:"data"`
}

func (vi *VaultService) Create(ctx context.Context, userID int64, item VaultCreate) (int64, error) {
	switch item.Type {
	case model.ItemLogin:
		var data LoginPasswordData
		err := json.Unmarshal(item.Data, &data)
		if err != nil {
			return 0, err
		}

		if err := data.validate(); err != nil {
			return 0, err
		}

	case model.ItemBankCard:
		var data BankCardData
		err := json.Unmarshal(item.Data, &data)
		if err != nil {
			return 0, err
		}

		if err := data.validate(); err != nil {
			return 0, err
		}

	case model.ItemText:
		var data TextData
		err := json.Unmarshal(item.Data, &data)
		if err != nil {
			return 0, err
		}

		if err := data.validate(); err != nil {
			return 0, err
		}
	case model.ItemBinary:
		var data FileData

		err := json.Unmarshal(item.Data, &data)
		if err != nil {
			return 0, err
		}

		if err := data.validate(); err != nil {
			return 0, err
		}

	default:
		return 0, errors.New("unknown vault item type")
	}

	vaultItem := model.VaultItem{
		UserID:     userID,
		Type:       item.Type,
		Title:      item.Title,
		SecretData: item.Data,
	}

	newItemId, err := vi.vaultRepo.Create(ctx, vaultItem)

	if err != nil {
		return 0, err
	}

	err = vi.vaultRepo.CreateMetadata(ctx, newItemId, item.Metadata)

	return newItemId, err
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

func (fd *FileData) validate() error {
	if fd.Name == "" || len(fd.Data) == 0 {
		return errors.New("file is required")
	}
	return nil
}

func (vi *VaultService) Get(ctx context.Context, itemID, userID int64) (dto.VaultResponseGet, error) {
	vault, err := vi.vaultRepo.GetByID(ctx, itemID, userID)
	if err != nil {
		return dto.VaultResponseGet{}, err
	}

	metadata, err := vi.vaultRepo.GetMetadata(ctx, itemID)
	if err != nil {
		return dto.VaultResponseGet{}, err
	}

	result := dto.VaultResponseGet{
		ID:        vault.ID,
		Type:      string(vault.Type),
		Title:     vault.Title,
		Data:      vault.SecretData,
		Metadata:  metadata,
		CreatedAt: vault.CreatedAt,
		UpdatedAt: vault.UpdatedAt,
	}

	return result, nil
}

func (vi *VaultService) List(ctx context.Context, userID int64) ([]dto.VaultResponseGet, error) {
	vaults, err := vi.vaultRepo.List(ctx, userID)
	if err != nil {
		return []dto.VaultResponseGet{}, err
	}

	var result []dto.VaultResponseGet

	for _, vault := range vaults {
		metadata, err := vi.vaultRepo.GetMetadata(ctx, vault.ID)
		if err != nil {
			return []dto.VaultResponseGet{}, err
		}
		result = append(result, dto.VaultResponseGet{
			ID:        vault.ID,
			Type:      string(vault.Type),
			Title:     vault.Title,
			Data:      vault.SecretData,
			Metadata:  metadata,
			CreatedAt: vault.CreatedAt,
			UpdatedAt: vault.UpdatedAt,
		})
	}

	return result, nil
}
