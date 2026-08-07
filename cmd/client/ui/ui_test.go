package ui

import (
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"passwordmanager/cmd/client/api"
	service_model "passwordmanager/internal/model"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func key(k string) tea.KeyPressMsg {
	return tea.KeyPressMsg{
		Text: k,
	}
}

func TestInitialModel(t *testing.T) {
	m := InitialModel(nil)

	assert.Equal(
		t,
		MenuScreen,
		m.screen,
	)

	assert.Equal(
		t,
		[]string{
			"Login",
			"Register",
			"Exit",
		},
		m.items,
	)

	assert.Equal(
		t,
		0,
		m.cursor,
	)
}

func TestParseMetadataInput(t *testing.T) {
	result := parseMetadataInput(
		"env=prod; owner=test; empty=; invalid",
	)

	assert.Equal(
		t,
		map[string]string{
			"env":     "prod",
			"owner":   "test",
			"empty":   "",
			"invalid": "",
		},
		result,
	)
}

func TestFormatMetadata(t *testing.T) {
	result := formatMetadata(
		map[string]string{
			"z": "3",
			"a": "1",
		},
	)

	assert.Equal(
		t,
		"a=1, z=3",
		result,
	)
}

func TestFormatMetadata_Empty(t *testing.T) {
	assert.Equal(
		t,
		"-",
		formatMetadata(nil),
	)
}

func TestBuildVaultData_Login(t *testing.T) {
	m := InitialModel(nil)

	m.selectedType = service_model.ItemLogin

	m.vaultLogin.SetValue("admin")
	m.vaultPassword.SetValue("secret")

	result := buildVaultData(m)

	assert.Equal(
		t,
		map[string]any{
			"login":    "admin",
			"password": "secret",
		},
		result,
	)
}

func TestDecodeSecretData_Login(t *testing.T) {
	raw := `{
		"login":"admin",
		"password":"secret"
	}`

	encoded := base64.StdEncoding.EncodeToString(
		[]byte(raw),
	)

	result, err := decodeSecretData(
		string(service_model.ItemLogin),
		encoded,
	)

	require.NoError(t, err)

	assert.Equal(
		t,
		"login: admin | password: secret",
		result,
	)
}

func TestDecodeSecretData_InvalidBase64(t *testing.T) {
	_, err := decodeSecretData(
		string(service_model.ItemLogin),
		"broken",
	)

	assert.Error(t, err)
}

func TestVaultTypeNavigation(t *testing.T) {
	m := InitialModel(nil)

	m.screen = VaultTypeScreen

	updated, _ := vaultTypeUpdate(
		m,
		key("down"),
	)

	result := updated.(model)

	assert.Equal(
		t,
		1,
		result.cursor,
	)
}

func TestMenuEnterLogin(t *testing.T) {
	m := InitialModel(nil)

	updated, _ := menuUpdate(
		m,
		key("enter"),
	)

	result := updated.(model)

	assert.Equal(
		t,
		LoginScreen,
		result.screen,
	)
}

func TestSaveBinaryFileCmd(t *testing.T) {
	raw := []byte("hello")

	_ = base64.StdEncoding.EncodeToString(raw)

	item := api.VaultItem{
		ID: 1,
		Data: base64.StdEncoding.EncodeToString(
			[]byte(`{
				"name":"test.txt",
				"data":"aGVsbG8="
			}`),
		),
	}

	cmd := saveBinaryFileCmd(item)

	msg := cmd()

	result := msg.(VaultSaveResultMsg)

	require.NoError(t, result.err)

	defer os.Remove(
		filepath.Base(result.path),
	)
}

func TestBuildVaultData_Text(t *testing.T) {
	m := InitialModel(nil)

	m.selectedType = service_model.ItemText
	m.vaultText.SetValue("hello")

	result := buildVaultData(m)

	assert.Equal(
		t,
		map[string]any{
			"text": "hello",
		},
		result,
	)
}

func TestBuildVaultData_BankCard(t *testing.T) {
	m := InitialModel(nil)

	m.selectedType = service_model.ItemBankCard

	m.vaultNumber.SetValue("1234")
	m.vaultHolder.SetValue("John")
	m.vaultMonth.SetValue("12")
	m.vaultYear.SetValue("2030")
	m.vaultCVV.SetValue("777")

	result := buildVaultData(m)

	assert.Equal(
		t,
		map[string]any{
			"number": "1234",
			"holder": "John",
			"month":  "12",
			"year":   "2030",
			"cvv":    "777",
		},
		result,
	)
}

func TestDecodeSecretData_Text(t *testing.T) {
	raw := `{
		"text":"hello"
	}`

	encoded := base64.StdEncoding.EncodeToString(
		[]byte(raw),
	)

	result, err := decodeSecretData(
		string(service_model.ItemText),
		encoded,
	)

	require.NoError(t, err)

	assert.Equal(
		t,
		"text: hello",
		result,
	)
}

func TestDecodeSecretData_BankCard(t *testing.T) {
	raw := `{
		"number":"123",
		"holder":"John",
		"month":5,
		"year":2030,
		"cvv":"123"
	}`

	encoded := base64.StdEncoding.EncodeToString(
		[]byte(raw),
	)

	result, err := decodeSecretData(
		string(service_model.ItemBankCard),
		encoded,
	)

	require.NoError(t, err)

	assert.Equal(
		t,
		"number: 123 | holder: John | exp: 05/2030 | cvv: 123",
		result,
	)
}

func TestVaultTypeNavigationUpDown(t *testing.T) {
	m := InitialModel(nil)
	m.screen = VaultTypeScreen

	updated, _ := vaultTypeUpdate(
		m,
		key("down"),
	)

	result := updated.(model)

	assert.Equal(t, 1, result.cursor)

	updated, _ = vaultTypeUpdate(
		result,
		key("up"),
	)

	result = updated.(model)

	assert.Equal(t, 0, result.cursor)
}

func TestVaultTypeSelect(t *testing.T) {
	m := InitialModel(nil)

	m.screen = VaultTypeScreen

	updated, _ := vaultTypeUpdate(
		m,
		key("enter"),
	)

	result := updated.(model)

	assert.Equal(
		t,
		VaultCreateScreen,
		result.screen,
	)
}

func TestMenuNavigateRegister(t *testing.T) {
	m := InitialModel(nil)

	updated, _ := menuUpdate(
		m,
		key("down"),
	)

	result := updated.(model)

	assert.Equal(
		t,
		1,
		result.cursor,
	)

	updated, _ = menuUpdate(
		result,
		key("enter"),
	)

	result = updated.(model)

	assert.Equal(
		t,
		RegisterScreen,
		result.screen,
	)
}

func TestMenuExit(t *testing.T) {
	m := InitialModel(nil)

	m.cursor = 2

	_, cmd := menuUpdate(
		m,
		key("enter"),
	)

	assert.NotNil(
		t,
		cmd,
	)
}

func TestLoginResultSuccess(t *testing.T) {
	m := InitialModel(nil)

	updated, _ := m.Update(
		LoginResultMsg{},
	)

	result := updated.(model)

	assert.Equal(
		t,
		VaultScreen,
		result.screen,
	)

	assert.Empty(
		t,
		result.loginError,
	)
}

func TestLoginResultError(t *testing.T) {
	m := InitialModel(nil)

	updated, _ := m.Update(
		LoginResultMsg{
			err: errors.New("wrong password"),
		},
	)

	result := updated.(model)

	assert.Equal(
		t,
		"wrong password",
		result.loginError,
	)
}

func TestRegisterResultSuccess(t *testing.T) {
	m := InitialModel(nil)

	m.regUsername.SetValue("admin")

	updated, _ := m.Update(
		RegisterResultMsg{},
	)

	result := updated.(model)

	assert.Equal(
		t,
		LoginScreen,
		result.screen,
	)

	assert.Equal(
		t,
		"Регистрация прошла успешно, теперь можно войти",
		result.registerOK,
	)
}

func TestVaultListResultSuccess(t *testing.T) {
	m := InitialModel(nil)

	items := []api.VaultItem{
		{
			ID:    1,
			Title: "github",
		},
	}

	updated, _ := m.Update(
		VaultListResultMsg{
			items: items,
		},
	)

	result := updated.(model)

	assert.Len(
		t,
		result.vaultItems,
		1,
	)
}

func TestVaultListResultError(t *testing.T) {
	m := InitialModel(nil)

	updated, _ := m.Update(
		VaultListResultMsg{
			err: errors.New("db error"),
		},
	)

	result := updated.(model)

	assert.Equal(
		t,
		"db error",
		result.vaultError,
	)
}

func TestCreateFieldsFor_Login(t *testing.T) {
	m := InitialModel(nil)

	m.selectedType = service_model.ItemLogin

	fields := createFieldsFor(&m)

	assert.Len(t, fields, 4)
}

func TestCreateFieldsFor_Text(t *testing.T) {
	m := InitialModel(nil)

	m.selectedType = service_model.ItemText

	fields := createFieldsFor(&m)

	assert.Len(t, fields, 3)
}

func TestCreateFieldsFor_BankCard(t *testing.T) {
	m := InitialModel(nil)

	m.selectedType = service_model.ItemBankCard

	fields := createFieldsFor(&m)

	assert.Len(t, fields, 7)
}

func TestCreateLabelsFor_All(t *testing.T) {

	tests := []struct {
		name string
		typ  service_model.ItemType
		size int
	}{
		{
			name: "login",
			typ:  service_model.ItemLogin,
			size: 4,
		},
		{
			name: "text",
			typ:  service_model.ItemText,
			size: 3,
		},
		{
			name: "binary",
			typ:  service_model.ItemBinary,
			size: 3,
		},
		{
			name: "card",
			typ:  service_model.ItemBankCard,
			size: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			result := createLabelsFor(tt.typ)

			assert.Len(
				t,
				result,
				tt.size,
			)
		})
	}
}

func TestDecodeSecretData_Card(t *testing.T) {

	raw := `{
		"number":"1111",
		"holder":"John",
		"month":10,
		"year":2030,
		"cvv":"123"
	}`

	result, err := decodeSecretData(
		string(service_model.ItemBankCard),
		encode(raw),
	)

	require.NoError(t, err)

	assert.Contains(
		t,
		result,
		"1111",
	)
}

func TestDecodeSecretData_Binary(t *testing.T) {

	raw := `{
		"name":"file.txt",
		"data":"aGVsbG8="
	}`

	result, err := decodeSecretData(
		string(service_model.ItemBinary),
		encode(raw),
	)

	require.NoError(t, err)

	assert.Contains(
		t,
		result,
		"file.txt",
	)
}

func TestDecodeSecretData_InvalidJSON(t *testing.T) {

	_, err := decodeSecretData(
		string(service_model.ItemText),
		encode("broken"),
	)

	assert.Error(t, err)
}

func TestMenuNavigation(t *testing.T) {

	m := InitialModel(nil)

	m.cursor = 0

	result, _ := menuUpdate(
		m,
		key("down"),
	)

	assert.Equal(
		t,
		1,
		result.(model).cursor,
	)

	result, _ = menuUpdate(
		result.(model),
		key("up"),
	)

	assert.Equal(
		t,
		0,
		result.(model).cursor,
	)
}

func TestVaultTypeEsc(t *testing.T) {

	m := InitialModel(nil)

	m.screen = VaultTypeScreen

	result, _ := vaultTypeUpdate(
		m,
		key("esc"),
	)

	assert.Equal(
		t,
		VaultScreen,
		result.(model).screen,
	)
}

func TestLoginErrorMessage(t *testing.T) {

	m := InitialModel(nil)

	result, _ := m.Update(
		LoginResultMsg{
			err: errors.New("bad login"),
		},
	)

	got := result.(model)

	assert.Equal(
		t,
		"bad login",
		got.loginError,
	)
}

func TestRegisterSuccess(t *testing.T) {

	m := InitialModel(nil)

	m.regUsername.SetValue("admin")

	result, _ := m.Update(
		RegisterResultMsg{},
	)

	got := result.(model)

	assert.Equal(
		t,
		LoginScreen,
		got.screen,
	)

	assert.Equal(
		t,
		"admin",
		got.username.Value(),
	)
}

func TestVaultListResult(t *testing.T) {

	m := InitialModel(nil)

	result, _ := m.Update(
		VaultListResultMsg{
			items: []api.VaultItem{
				{
					ID:    1,
					Title: "github",
				},
			},
		},
	)

	got := result.(model)

	assert.Len(
		t,
		got.vaultItems,
		1,
	)
}

func TestVaultDetailResult(t *testing.T) {

	m := InitialModel(nil)

	result, _ := m.Update(
		VaultItemResultMsg{
			item: api.VaultItem{
				ID: 1,
			},
		},
	)

	got := result.(model)

	assert.Equal(
		t,
		VaultDetailScreen,
		got.screen,
	)
}

func TestViews(t *testing.T) {

	m := InitialModel(nil)

	tests := []Screen{
		MenuScreen,
		LoginScreen,
		RegisterScreen,
		VaultScreen,
		VaultListScreen,
		VaultTypeScreen,
		VaultCreateScreen,
		VaultDetailScreen,
	}

	for _, screen := range tests {

		m.screen = screen

		view := m.View()

		assert.NotEmpty(
			t,
			view.Content,
		)
	}
}

func encode(s string) string {
	return base64.StdEncoding.EncodeToString(
		[]byte(s),
	)
}

func TestLoginUpdateEnter(t *testing.T) {
	m := InitialModel(nil)

	m.screen = LoginScreen

	m.username.SetValue("admin")
	m.password.SetValue("123")

	m.username.Focus()

	updated, cmd := loginUpdate(
		m,
		key("enter"),
	)

	assert.NotNil(t, cmd)

	_ = updated
}

func TestLoginUpdateEsc(t *testing.T) {
	m := InitialModel(nil)

	m.screen = LoginScreen

	updated, _ := loginUpdate(
		m,
		key("esc"),
	)

	assert.Equal(
		t,
		MenuScreen,
		updated.(model).screen,
	)
}

func TestRegisterUpdateEmpty(t *testing.T) {
	m := InitialModel(nil)

	m.screen = RegisterScreen

	updated, _ := registerUpdate(
		m,
		key("enter"),
	)

	result := updated.(model)

	assert.Equal(
		t,
		"login and password are required",
		result.registerError,
	)
}

func TestRegisterUpdateEsc(t *testing.T) {
	m := InitialModel(nil)

	m.screen = RegisterScreen

	updated, _ := registerUpdate(
		m,
		key("esc"),
	)

	assert.Equal(
		t,
		MenuScreen,
		updated.(model).screen,
	)
}
func TestVaultUpdateNavigation(t *testing.T) {

	m := InitialModel(nil)

	m.screen = VaultScreen

	result, _ := vaultUpdate(
		m,
		key("down"),
	)

	assert.Equal(
		t,
		1,
		result.(model).cursor,
	)
}

func TestVaultUpdateEsc(t *testing.T) {

	m := InitialModel(nil)

	m.screen = VaultScreen

	result, _ := vaultUpdate(
		m,
		key("esc"),
	)

	assert.Equal(
		t,
		MenuScreen,
		result.(model).screen,
	)
}
func TestVaultDetailView(t *testing.T) {

	m := InitialModel(nil)

	m.screen = VaultDetailScreen

	m.vaultDetail = &api.VaultItem{
		ID:    1,
		Title: "github",
		Type:  string(service_model.ItemText),
		Data: encode(`{
			"text":"hello"
		}`),
		Metadata: map[string]string{
			"env": "prod",
		},
	}

	view := vaultDetailView(m)

	assert.Contains(
		t,
		view.Content,
		"github",
	)

	assert.Contains(
		t,
		view.Content,
		"hello",
	)
}

func TestVaultListViewEmpty(t *testing.T) {

	m := InitialModel(nil)

	m.screen = VaultListScreen

	view := vaultListView(m)

	assert.Contains(
		t,
		view.Content,
		"Vault is empty",
	)
}

func TestVaultListViewItems(t *testing.T) {

	m := InitialModel(nil)

	m.vaultItems = []api.VaultItem{
		{
			ID:    1,
			Title: "github",
			Type:  "LOGIN",
		},
	}

	view := vaultListView(m)

	assert.Contains(
		t,
		view.Content,
		"github",
	)
}

func TestUpdateMessages(t *testing.T) {

	m := InitialModel(nil)

	tests := []tea.Msg{
		LoginResultMsg{},
		RegisterResultMsg{},
		VaultListResultMsg{},
		VaultItemResultMsg{},
		VaultSaveResultMsg{},
		VaultCreateResultMsg{},
	}

	for _, msg := range tests {

		result, _ := m.Update(msg)

		assert.NotNil(
			t,
			result,
		)
	}
}
