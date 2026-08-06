package service

import (
	"context"
	"log/slog"
	"passwordmanager/internal/model"
	"passwordmanager/internal/repository"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockVaultRepository struct {
	mock.Mock
}

func (m *MockVaultRepository) Create(
	ctx context.Context,
	item model.VaultItem,
) (int64, error) {
	args := m.Called(ctx, item)

	return args.Get(0).(int64), args.Error(1)
}

func (m *MockVaultRepository) CreateMetadata(
	ctx context.Context,
	itemID int64,
	metadata map[string]string,
) error {
	args := m.Called(ctx, itemID, metadata)

	return args.Error(0)
}

func (m *MockVaultRepository) GetByID(
	ctx context.Context,
	id int64,
	userID int64,
) (*model.VaultItem, error) {
	args := m.Called(ctx, id, userID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*model.VaultItem), args.Error(1)
}

func (m *MockVaultRepository) GetMetadata(
	ctx context.Context,
	itemID int64,
) (map[string]string, error) {
	args := m.Called(ctx, itemID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockVaultRepository) List(
	ctx context.Context,
	userID int64,
) ([]model.VaultItem, error) {
	args := m.Called(ctx, userID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]model.VaultItem), args.Error(1)
}

var _ repository.VaultRepositoryInterface = (*MockVaultRepository)(nil)

func TestVaultService_Get(t *testing.T) {
	repo := new(MockVaultRepository)

	service := NewVaultService(
		slog.Default(),
		repo,
	)

	repo.On(
		"GetByID",
		mock.Anything,
		int64(1),
		int64(2),
	).
		Return(
			&model.VaultItem{
				ID:    1,
				Title: "test",
			},
			nil,
		)

	repo.On(
		"GetMetadata",
		mock.Anything,
		int64(1),
	).
		Return(
			map[string]string{
				"a": "b",
			},
			nil,
		)

	result, err := service.Get(
		context.Background(),
		1,
		2,
	)

	require.NoError(t, err)
	require.Equal(t, "test", result.Title)
	require.Equal(
		t,
		map[string]string{"a": "b"},
		result.Metadata,
	)

	repo.AssertExpectations(t)
}
func TestVaultService_Create_Login_Success(t *testing.T) {
	repo := new(MockVaultRepository)

	service := NewVaultService(
		slog.Default(),
		repo,
	)

	data := []byte(`{
		"login":"admin",
		"password":"123"
	}`)

	repo.
		On(
			"Create",
			mock.Anything,
			mock.Anything,
		).
		Return(
			int64(1),
			nil,
		)

	repo.
		On(
			"CreateMetadata",
			mock.Anything,
			int64(1),
			map[string]string{"env": "prod"},
		).
		Return(nil)

	id, err := service.Create(
		context.Background(),
		10,
		VaultCreate{
			Type:  model.ItemLogin,
			Title: "test",
			Data:  data,
			Metadata: map[string]string{
				"env": "prod",
			},
		},
	)

	require.NoError(t, err)
	require.Equal(t, int64(1), id)
}

func TestTextDataValidate(t *testing.T) {
	valid := TextData{
		Text: "hello",
	}

	require.NoError(
		t,
		valid.validate(),
	)

	empty := TextData{}

	require.Error(
		t,
		empty.validate(),
	)
}

func TestFileDataValidate(t *testing.T) {
	valid := FileData{
		Name: "test.txt",
		Data: []byte("abc"),
	}

	require.NoError(
		t,
		valid.validate(),
	)

	empty := FileData{}

	require.Error(
		t,
		empty.validate(),
	)
}

func TestBankCardValidate(t *testing.T) {
	valid := BankCardData{
		Number: "1111",
		Holder: "John",
		Month:  1,
		Year:   2026,
		CVV:    "123",
	}

	require.NoError(
		t,
		valid.validate(),
	)

	empty := BankCardData{}

	require.Error(
		t,
		empty.validate(),
	)
}
