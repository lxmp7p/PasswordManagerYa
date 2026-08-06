package service

import (
	"context"
	"errors"
	"log/slog"
	"passwordmanager/internal/model"
	"passwordmanager/internal/repository"
	"testing"
	"time"

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

func TestVaultService_List_Success(t *testing.T) {
	repo := new(MockVaultRepository)

	service := NewVaultService(
		slog.Default(),
		repo,
	)

	now := time.Now()

	repo.On(
		"List",
		mock.Anything,
		int64(2),
	).
		Return(
			[]model.VaultItem{
				{
					ID:         1,
					UserID:     2,
					Type:       model.ItemLogin,
					Title:      "github",
					SecretData: []byte("secret"),
					CreatedAt:  now,
					UpdatedAt:  now,
				},
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
				"login": "admin",
			},
			nil,
		)

	result, err := service.List(
		context.Background(),
		2,
	)

	require.NoError(t, err)
	require.Len(t, result, 1)

	require.Equal(t, int64(1), result[0].ID)
	require.Equal(t, "github", result[0].Title)
	require.Equal(
		t,
		map[string]string{"login": "admin"},
		result[0].Metadata,
	)

	repo.AssertExpectations(t)
}

func TestVaultService_List_Error(t *testing.T) {
	repo := new(MockVaultRepository)

	service := NewVaultService(
		slog.Default(),
		repo,
	)

	repo.On(
		"List",
		mock.Anything,
		int64(2),
	).
		Return(
			nil,
			errors.New("repository error"),
		)

	result, err := service.List(
		context.Background(),
		2,
	)

	require.Error(t, err)
	require.Empty(t, result)

	repo.AssertExpectations(t)
}

func TestVaultService_List_MetadataError(t *testing.T) {
	repo := new(MockVaultRepository)

	service := NewVaultService(
		slog.Default(),
		repo,
	)

	repo.On(
		"List",
		mock.Anything,
		int64(2),
	).
		Return(
			[]model.VaultItem{
				{
					ID: 1,
				},
			},
			nil,
		)

	repo.On(
		"GetMetadata",
		mock.Anything,
		int64(1),
	).
		Return(
			nil,
			errors.New("metadata error"),
		)

	result, err := service.List(
		context.Background(),
		2,
	)

	require.Error(t, err)
	require.Empty(t, result)

	repo.AssertExpectations(t)
}
func TestVaultService_Create_CreateError(t *testing.T) {
	repo := new(MockVaultRepository)

	service := NewVaultService(
		slog.Default(),
		repo,
	)

	repo.On(
		"Create",
		mock.Anything,
		mock.Anything,
	).
		Return(
			int64(0),
			errors.New("create error"),
		)

	id, err := service.Create(
		context.Background(),
		10,
		VaultCreate{
			Type:  model.ItemLogin,
			Title: "test",
			Data: []byte(`{
				"login":"admin",
				"password":"123"
			}`),
		},
	)

	require.Error(t, err)
	require.Equal(t, int64(0), id)

	repo.AssertExpectations(t)
}

func TestVaultService_Create_MetadataError(t *testing.T) {
	repo := new(MockVaultRepository)

	service := NewVaultService(
		slog.Default(),
		repo,
	)

	repo.On(
		"Create",
		mock.Anything,
		mock.Anything,
	).
		Return(
			int64(1),
			nil,
		)

	repo.On(
		"CreateMetadata",
		mock.Anything,
		int64(1),
		mock.Anything,
	).
		Return(
			errors.New("metadata error"),
		)

	id, err := service.Create(
		context.Background(),
		10,
		VaultCreate{
			Type:  model.ItemLogin,
			Title: "test",
			Data: []byte(`{
				"login":"admin",
				"password":"123"
			}`),
			Metadata: map[string]string{
				"env": "prod",
			},
		},
	)

	require.NoError(t, err)
	require.Equal(t, int64(1), id)

	repo.AssertExpectations(t)
}

func TestVaultService_Create_UnknownType(t *testing.T) {
	repo := new(MockVaultRepository)

	service := NewVaultService(
		slog.Default(),
		repo,
	)

	id, err := service.Create(
		context.Background(),
		1,
		VaultCreate{
			Type: "unknown",
		},
	)

	require.Error(t, err)
	require.Equal(t, int64(0), id)
}

func TestVaultService_Create_Text_Success(t *testing.T) {
	repo := new(MockVaultRepository)

	service := NewVaultService(
		slog.Default(),
		repo,
	)

	repo.On(
		"Create",
		mock.Anything,
		mock.Anything,
	).
		Return(
			int64(1),
			nil,
		)

	repo.On(
		"CreateMetadata",
		mock.Anything,
		int64(1),
		mock.Anything,
	).
		Return(nil)

	id, err := service.Create(
		context.Background(),
		10,
		VaultCreate{
			Type:  model.ItemText,
			Title: "note",
			Data: []byte(`{
				"text":"hello"
			}`),
		},
	)

	require.NoError(t, err)
	require.Equal(t, int64(1), id)

	repo.AssertExpectations(t)
}

func TestVaultService_Create_BankCard_Success(t *testing.T) {
	repo := new(MockVaultRepository)

	service := NewVaultService(
		slog.Default(),
		repo,
	)

	repo.On(
		"Create",
		mock.Anything,
		mock.Anything,
	).
		Return(
			int64(1),
			nil,
		)

	repo.On(
		"CreateMetadata",
		mock.Anything,
		int64(1),
		mock.Anything,
	).
		Return(nil)

	id, err := service.Create(
		context.Background(),
		10,
		VaultCreate{
			Type:  model.ItemBankCard,
			Title: "card",
			Data: []byte(`{
				"number":"1111",
				"holder":"John Doe",
				"month":1,
				"year":2026,
				"cvv":"123"
			}`),
		},
	)

	require.NoError(t, err)
	require.Equal(t, int64(1), id)

	repo.AssertExpectations(t)
}

func TestVaultService_Create_Binary_Success(t *testing.T) {
	repo := new(MockVaultRepository)

	service := NewVaultService(
		slog.Default(),
		repo,
	)

	repo.On(
		"Create",
		mock.Anything,
		mock.Anything,
	).
		Return(
			int64(1),
			nil,
		)

	repo.On(
		"CreateMetadata",
		mock.Anything,
		int64(1),
		mock.Anything,
	).
		Return(nil)

	id, err := service.Create(
		context.Background(),
		10,
		VaultCreate{
			Type:  model.ItemBinary,
			Title: "file",
			Data: []byte(`{
				"name":"test.txt",
				"data":"YWJj"
			}`),
		},
	)

	require.NoError(t, err)
	require.Equal(t, int64(1), id)

	repo.AssertExpectations(t)
}

func TestVaultService_Get_GetByIDError(t *testing.T) {
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
			nil,
			errors.New("not found"),
		)

	result, err := service.Get(
		context.Background(),
		1,
		2,
	)

	require.Error(t, err)
	require.Empty(t, result)

	repo.AssertExpectations(t)
}

func TestVaultService_Get_MetadataError(t *testing.T) {
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
				ID: 1,
			},
			nil,
		)

	repo.On(
		"GetMetadata",
		mock.Anything,
		int64(1),
	).
		Return(
			nil,
			errors.New("metadata error"),
		)

	result, err := service.Get(
		context.Background(),
		1,
		2,
	)

	require.Error(t, err)
	require.Empty(t, result)

	repo.AssertExpectations(t)
}

func TestVaultService_Create_InvalidJSON(t *testing.T) {
	repo := new(MockVaultRepository)

	service := NewVaultService(
		slog.Default(),
		repo,
	)

	id, err := service.Create(
		context.Background(),
		1,
		VaultCreate{
			Type: model.ItemLogin,
			Data: []byte(`{invalid}`),
		},
	)

	require.Error(t, err)
	require.Equal(t, int64(0), id)
}

func TestVaultService_Create_LoginValidationError(t *testing.T) {
	repo := new(MockVaultRepository)

	service := NewVaultService(
		slog.Default(),
		repo,
	)

	id, err := service.Create(
		context.Background(),
		1,
		VaultCreate{
			Type: model.ItemLogin,
			Data: []byte(`{
				"login":"",
				"password":""
			}`),
		},
	)

	require.Error(t, err)
	require.Equal(t, int64(0), id)
}
