package ui

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"passwordmanager/cmd/client/api"
	service_model "passwordmanager/internal/model"
	"path/filepath"
	"sort"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type Client struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

type MenuItem struct {
	Name   string
	Action func() tea.Cmd
}

type Screen int

const (
	MenuScreen Screen = iota
	LoginScreen
	RegisterScreen
	VaultScreen
	VaultListScreen
	VaultTypeScreen   // выбор типа нового vault-элемента
	VaultCreateScreen // форма ввода данных под выбранный тип
	VaultDetailScreen // детальный просмотр одного элемента (после GetVaultByID)
)

// itemTypes — порядок отображения в экране выбора типа.
var itemTypes = []service_model.ItemType{
	service_model.ItemLogin,
	service_model.ItemText,
	service_model.ItemBinary,
	service_model.ItemBankCard,
}

// itemTypeLabels — человекочитаемые подписи для itemTypes (индексы совпадают).
var itemTypeLabels = []string{
	"Логин / пароль",
	"Текст",
	"Бинарные данные",
	"Банковская карта",
}

type model struct {
	screen Screen

	client *api.Client

	cursor int
	items  []string

	username textinput.Model
	password textinput.Model

	loginError string

	regUsername   textinput.Model
	regPassword   textinput.Model
	registerError string
	registerOK    string

	vaultItems  []api.VaultItem
	vaultError  string
	vaultStatus string

	vaultDetail *api.VaultItem

	selectedType service_model.ItemType

	createTitle textinput.Model
	createFocus int

	vaultLogin    textinput.Model
	vaultPassword textinput.Model
	vaultText     textinput.Model
	vaultBinary   textinput.Model
	vaultNumber   textinput.Model
	vaultHolder   textinput.Model
	vaultMonth    textinput.Model
	vaultYear     textinput.Model
	vaultCVV      textinput.Model

	vaultMetadata textinput.Model
}

type LoginResultMsg struct {
	err error
}

type RegisterResultMsg struct {
	err error
}

type VaultListResultMsg struct {
	items []api.VaultItem
	err   error
}

type VaultCreateResultMsg struct {
	err error
}

type VaultItemResultMsg struct {
	item api.VaultItem
	err  error
}

type VaultSaveResultMsg struct {
	path string
	err  error
}

func InitialModel(client *api.Client) model {
	username := textinput.New()
	username.Placeholder = "username"

	password := textinput.New()
	password.Placeholder = "password"
	password.EchoMode = textinput.EchoPassword

	regUsername := textinput.New()
	regUsername.Placeholder = "username"

	regPassword := textinput.New()
	regPassword.Placeholder = "password"
	regPassword.EchoMode = textinput.EchoPassword

	createTitle := textinput.New()
	createTitle.Placeholder = "title"

	vaultLogin := textinput.New()
	vaultLogin.Placeholder = "login"

	vaultPassword := textinput.New()
	vaultPassword.Placeholder = "password"

	vaultText := textinput.New()
	vaultText.Placeholder = "text"

	vaultBinary := textinput.New()
	vaultBinary.Placeholder = "path to file, e.g. /home/user/photo.png"

	vaultNumber := textinput.New()
	vaultNumber.Placeholder = "card number"

	vaultHolder := textinput.New()
	vaultHolder.Placeholder = "holder"

	vaultMonth := textinput.New()
	vaultMonth.Placeholder = "month"

	vaultYear := textinput.New()
	vaultYear.Placeholder = "year"

	vaultCVV := textinput.New()
	vaultCVV.Placeholder = "cvv"

	vaultMetadata := textinput.New()
	vaultMetadata.Placeholder = "key1=value1;key2=value2 (optional)"

	return model{
		screen: MenuScreen,
		client: client,

		items: []string{
			"Login",
			"Register",
			"Exit",
		},

		username:      username,
		password:      password,
		regUsername:   regUsername,
		regPassword:   regPassword,
		createTitle:   createTitle,
		vaultLogin:    vaultLogin,
		vaultPassword: vaultPassword,
		vaultText:     vaultText,
		vaultBinary:   vaultBinary,
		vaultNumber:   vaultNumber,
		vaultHolder:   vaultHolder,
		vaultMonth:    vaultMonth,
		vaultYear:     vaultYear,
		vaultCVV:      vaultCVV,
		vaultMetadata: vaultMetadata,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case LoginResultMsg:
		if msg.err != nil {
			m.loginError = msg.err.Error()
			return m, nil
		}

		m.loginError = ""
		m.screen = VaultScreen
		m.cursor = 0
		m.vaultItems = nil

		return m, nil

	case RegisterResultMsg:
		if msg.err != nil {
			m.registerError = msg.err.Error()
			m.registerOK = ""
			return m, nil
		}

		m.registerError = ""
		m.registerOK = "Регистрация прошла успешно, теперь можно войти"
		m.screen = LoginScreen
		m.username.SetValue(m.regUsername.Value())
		m.username.Blur()
		m.password.SetValue("")
		m.password.Focus()
		m.regUsername.SetValue("")
		m.regPassword.SetValue("")

		return m, nil

	case VaultListResultMsg:
		if msg.err != nil {
			m.vaultError = msg.err.Error()
			return m, nil
		}

		m.vaultItems = msg.items
		m.vaultError = ""

		return m, nil

	case VaultItemResultMsg:
		if msg.err != nil {
			m.vaultError = msg.err.Error()
			return m, nil
		}

		item := msg.item
		m.vaultDetail = &item
		m.screen = VaultDetailScreen
		m.vaultError = ""
		m.vaultStatus = ""

		return m, nil

	case VaultSaveResultMsg:
		if msg.err != nil {
			m.vaultError = msg.err.Error()
			m.vaultStatus = ""
			return m, nil
		}

		m.vaultError = ""
		m.vaultStatus = "Сохранено: " + msg.path

		return m, nil

	case VaultCreateResultMsg:
		if msg.err != nil {
			m.vaultError = msg.err.Error()
			return m, nil
		}

		m.screen = VaultListScreen
		m.cursor = 0

		return m, loadVaultsCmd(m.client)
	}

	switch m.screen {

	case MenuScreen:
		return menuUpdate(m, msg)

	case LoginScreen:
		return loginUpdate(m, msg)

	case RegisterScreen:
		return registerUpdate(m, msg)

	case VaultScreen:
		return vaultUpdate(m, msg)

	case VaultListScreen:
		return vaultListUpdate(m, msg)

	case VaultDetailScreen:
		return vaultDetailUpdate(m, msg)

	case VaultTypeScreen:
		return vaultTypeUpdate(m, msg)

	case VaultCreateScreen:
		return vaultCreateUpdate(m, msg)
	}

	return m, nil
}

func createFieldsFor(m *model) []*textinput.Model {
	fields := []*textinput.Model{&m.createTitle}

	switch m.selectedType {
	case service_model.ItemLogin:
		fields = append(fields, &m.vaultLogin, &m.vaultPassword)
	case service_model.ItemText:
		fields = append(fields, &m.vaultText)
	case service_model.ItemBinary:
		fields = append(fields, &m.vaultBinary)
	case service_model.ItemBankCard:
		fields = append(fields, &m.vaultNumber, &m.vaultHolder, &m.vaultMonth, &m.vaultYear, &m.vaultCVV)
	}

	fields = append(fields, &m.vaultMetadata)

	return fields
}

func createLabelsFor(t service_model.ItemType) []string {
	labels := []string{"Title"}

	switch t {
	case service_model.ItemLogin:
		labels = append(labels, "Login", "Password")
	case service_model.ItemText:
		labels = append(labels, "Text")
	case service_model.ItemBinary:
		labels = append(labels, "Path to file")
	case service_model.ItemBankCard:
		labels = append(labels, "Card number", "Holder", "Month", "Year", "CVV")
	}

	labels = append(labels, "Metadata (key=value;key=value)")

	return labels
}

func resetCreateForm(m *model) {
	m.createTitle.SetValue("")
	m.vaultLogin.SetValue("")
	m.vaultPassword.SetValue("")
	m.vaultText.SetValue("")
	m.vaultBinary.SetValue("")
	m.vaultNumber.SetValue("")
	m.vaultHolder.SetValue("")
	m.vaultMonth.SetValue("")
	m.vaultYear.SetValue("")
	m.vaultCVV.SetValue("")
	m.vaultMetadata.SetValue("")

	m.vaultLogin.Blur()
	m.vaultPassword.Blur()
	m.vaultText.Blur()
	m.vaultBinary.Blur()
	m.vaultNumber.Blur()
	m.vaultHolder.Blur()
	m.vaultMonth.Blur()
	m.vaultYear.Blur()
	m.vaultCVV.Blur()
	m.vaultMetadata.Blur()

	m.createFocus = 0
	m.createTitle.Focus()
}

func buildVaultData(m model) map[string]any {
	data := map[string]any{}

	switch m.selectedType {
	case service_model.ItemLogin:
		data["login"] = m.vaultLogin.Value()
		data["password"] = m.vaultPassword.Value()
	case service_model.ItemText:
		data["text"] = m.vaultText.Value()
	case service_model.ItemBankCard:
		data["number"] = m.vaultNumber.Value()
		data["holder"] = m.vaultHolder.Value()
		data["month"] = m.vaultMonth.Value()
		data["year"] = m.vaultYear.Value()
		data["cvv"] = m.vaultCVV.Value()
	}

	return data
}

func parseMetadataInput(raw string) map[string]string {
	result := map[string]string{}

	for _, pair := range strings.Split(raw, ";") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		parts := strings.SplitN(pair, "=", 2)
		key := strings.TrimSpace(parts[0])
		if key == "" {
			continue
		}

		value := ""
		if len(parts) == 2 {
			value = strings.TrimSpace(parts[1])
		}

		result[key] = value
	}

	return result
}

func formatMetadata(md map[string]string) string {
	if len(md) == 0 {
		return "-"
	}

	keys := make([]string, 0, len(md))
	for k := range md {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, md[k]))
	}

	return strings.Join(pairs, ", ")
}

func createBinaryVaultCmd(client *api.Client, title, path string, metadata map[string]string) tea.Cmd {
	return func() tea.Msg {
		raw, err := os.ReadFile(path)
		if err != nil {
			return VaultCreateResultMsg{err: fmt.Errorf("не удалось прочитать файл: %w", err)}
		}

		item := api.VaultCreate{
			Title: title,
			Type:  string(service_model.ItemBinary),
			Data: map[string]any{
				"name": filepath.Base(path),
				"data": base64.StdEncoding.EncodeToString(raw),
			},
			Metadata: metadata,
		}

		err = client.CreateVault(item)
		return VaultCreateResultMsg{err: err}
	}
}

func decodeSecretData(itemType, encodedData string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return "", fmt.Errorf("не удалось раскодировать data: %w", err)
	}

	switch service_model.ItemType(itemType) {

	case service_model.ItemLogin:
		var d struct {
			Login    string `json:"login"`
			Password string `json:"password"`
		}
		if err := json.Unmarshal(raw, &d); err != nil {
			return "", err
		}
		return fmt.Sprintf("login: %s | password: %s", d.Login, d.Password), nil

	case service_model.ItemText:
		var d struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal(raw, &d); err != nil {
			return "", err
		}
		return fmt.Sprintf("text: %s", d.Text), nil

	case service_model.ItemBankCard:
		var d struct {
			Number string `json:"number"`
			Holder string `json:"holder"`
			Month  int    `json:"month"`
			Year   int    `json:"year"`
			CVV    string `json:"cvv"`
		}
		if err := json.Unmarshal(raw, &d); err != nil {
			return "", err
		}
		return fmt.Sprintf(
			"number: %s | holder: %s | exp: %02d/%d | cvv: %s",
			d.Number, d.Holder, d.Month, d.Year, d.CVV,
		), nil

	case service_model.ItemBinary:
		var d struct {
			Name string `json:"name"`
			Data []byte `json:"data"`
		}
		if err := json.Unmarshal(raw, &d); err != nil {
			return "", err
		}
		return fmt.Sprintf("file: %s (%d bytes)", d.Name, len(d.Data)), nil

	default:
		return string(raw), nil
	}
}

func saveBinaryFileCmd(item api.VaultItem) tea.Cmd {
	return func() tea.Msg {
		raw, err := base64.StdEncoding.DecodeString(item.Data)
		if err != nil {
			return VaultSaveResultMsg{err: fmt.Errorf("не удалось раскодировать data: %w", err)}
		}

		var d struct {
			Name string `json:"name"`
			Data []byte `json:"data"`
		}
		if err := json.Unmarshal(raw, &d); err != nil {
			return VaultSaveResultMsg{err: fmt.Errorf("не удалось разобрать file data: %w", err)}
		}

		outName := d.Name
		if outName == "" {
			outName = fmt.Sprintf("vault_%d.bin", item.ID)
		}

		if err := os.WriteFile(outName, d.Data, 0o600); err != nil {
			return VaultSaveResultMsg{err: fmt.Errorf("не удалось записать файл: %w", err)}
		}

		abs, err := filepath.Abs(outName)
		if err != nil {
			abs = outName
		}

		return VaultSaveResultMsg{path: abs}
	}
}

func vaultTypeUpdate(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyPressMsg:
		switch msg.String() {

		case "down":
			if m.cursor < len(itemTypes)-1 {
				m.cursor++
			}

		case "up":
			if m.cursor > 0 {
				m.cursor--
			}

		case "esc":
			m.screen = VaultScreen
			m.cursor = 0

		case "enter":
			m.selectedType = itemTypes[m.cursor]
			m.screen = VaultCreateScreen
			m.cursor = 0
			resetCreateForm(&m)
		}
	}

	return m, nil
}

func vaultCreateUpdate(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	fields := createFieldsFor(&m)

	var cmd tea.Cmd
	for _, f := range fields {
		if f.Focused() {
			*f, cmd = f.Update(msg)
			break
		}
	}

	switch msg := msg.(type) {

	case tea.KeyPressMsg:
		switch msg.String() {

		case "esc":
			m.screen = VaultTypeScreen

		case "tab":
			fields[m.createFocus].Blur()
			m.createFocus = (m.createFocus + 1) % len(fields)
			fields[m.createFocus].Focus()

		case "enter":
			metadata := parseMetadataInput(m.vaultMetadata.Value())

			if m.selectedType == service_model.ItemBinary {
				return m, createBinaryVaultCmd(m.client, m.createTitle.Value(), m.vaultBinary.Value(), metadata)
			}

			item := api.VaultCreate{
				Title:    m.createTitle.Value(),
				Type:     string(m.selectedType),
				Data:     buildVaultData(m),
				Metadata: metadata,
			}

			return m, func() tea.Msg {
				err := m.client.CreateVault(item)
				return VaultCreateResultMsg{err: err}
			}
		}
	}

	return m, cmd
}

func menuUpdate(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyPressMsg:
		switch msg.String() {

		case "down":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}

		case "up":
			if m.cursor > 0 {
				m.cursor--
			}

		case "enter":
			switch m.cursor {

			case 0:
				m.screen = LoginScreen
				m.username.Focus()
				m.password.Blur()
				m.cursor = 0
				m.registerOK = ""

			case 1:
				m.screen = RegisterScreen
				m.regUsername.Focus()
				m.regPassword.Blur()
				m.registerError = ""
				m.registerOK = ""

			case 2:
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func vaultUpdate(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyPressMsg:
		switch msg.String() {

		case "esc":
			m.screen = MenuScreen

		case "enter":
			switch m.cursor {

			case 0:
				m.screen = VaultListScreen
				m.cursor = 0
				return m, loadVaultsCmd(m.client)

			case 1:
				m.screen = VaultTypeScreen
				m.cursor = 0
				return m, nil

			case 2:
				m.screen = MenuScreen
			}

		case "down":
			if m.cursor < 2 {
				m.cursor++
			}

		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		}
	}

	return m, nil
}

func registerUpdate(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.regUsername.Focused() {
		m.regUsername, cmd = m.regUsername.Update(msg)
	}

	if m.regPassword.Focused() {
		m.regPassword, cmd = m.regPassword.Update(msg)
	}

	switch msg := msg.(type) {

	case tea.KeyPressMsg:
		switch msg.String() {

		case "tab":
			if m.regUsername.Focused() {
				m.regUsername.Blur()
				m.regPassword.Focus()
			} else {
				m.regPassword.Blur()
				m.regUsername.Focus()
			}

		case "esc":
			m.screen = MenuScreen

		case "enter":
			login := m.regUsername.Value()
			password := m.regPassword.Value()

			if login == "" || password == "" {
				m.registerError = "login and password are required"
				m.registerOK = ""
				return m, nil
			}

			return m, func() tea.Msg {
				err := m.client.Register(login, password)
				return RegisterResultMsg{err: err}
			}
		}
	}

	return m, cmd
}

func loginUpdate(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.username.Focused() {
		m.username, cmd = m.username.Update(msg)
	}

	if m.password.Focused() {
		m.password, cmd = m.password.Update(msg)
	}

	switch msg := msg.(type) {

	case tea.KeyPressMsg:
		switch msg.String() {

		case "tab":
			if m.username.Focused() {
				m.username.Blur()
				m.password.Focus()
			} else {
				m.password.Blur()
				m.username.Focus()
			}

		case "esc":
			m.screen = MenuScreen

		case "enter":
			return m, func() tea.Msg {
				err := m.client.Login(
					m.username.Value(),
					m.password.Value(),
				)

				return LoginResultMsg{err: err}
			}
		}
	}

	return m, cmd
}

func vaultListUpdate(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case VaultListResultMsg:
		if msg.err != nil {
			m.vaultError = msg.err.Error()
			return m, nil
		}

		m.vaultItems = msg.items
		m.vaultError = ""
		m.vaultStatus = ""
		if m.cursor >= len(m.vaultItems) {
			m.cursor = 0
		}

		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {

		case "esc":
			m.screen = VaultScreen
			m.vaultItems = nil
			m.vaultStatus = ""
			m.cursor = 0

		case "down":
			if m.cursor < len(m.vaultItems)-1 {
				m.cursor++
			}

		case "up":
			if m.cursor > 0 {
				m.cursor--
			}

		case "enter":
			if len(m.vaultItems) == 0 {
				return m, nil
			}

			return m, loadVaultById(m.client, m.vaultItems[m.cursor].ID)

		case "s":
			if len(m.vaultItems) == 0 {
				return m, nil
			}

			item := m.vaultItems[m.cursor]
			if item.Type != string(service_model.ItemBinary) {
				m.vaultError = "у этого элемента нет файла для сохранения"
				m.vaultStatus = ""
				return m, nil
			}

			return m, saveBinaryFileCmd(item)
		}
	}

	return m, nil
}

func vaultDetailUpdate(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyPressMsg:
		switch msg.String() {

		case "esc":
			m.screen = VaultListScreen
			m.vaultDetail = nil
			m.vaultError = ""
			m.vaultStatus = ""

		case "s":
			if m.vaultDetail == nil {
				return m, nil
			}

			if m.vaultDetail.Type != string(service_model.ItemBinary) {
				m.vaultError = "у этого элемента нет файла для сохранения"
				m.vaultStatus = ""
				return m, nil
			}

			return m, saveBinaryFileCmd(*m.vaultDetail)
		}
	}

	return m, nil
}

func loadVaultsCmd(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		items, err := client.ListVault()
		return VaultListResultMsg{items: items, err: err}
	}
}

func loadVaultById(client *api.Client, id int64) tea.Cmd {
	return func() tea.Msg {
		item, err := client.GetVaultByID(id)
		return VaultItemResultMsg{item: item, err: err}
	}
}

func (m model) View() tea.View {
	switch m.screen {

	case MenuScreen:
		return menuView(m)

	case LoginScreen:
		return loginView(m)

	case RegisterScreen:
		return registerView(m)

	case VaultScreen:
		return vaultView(m)

	case VaultListScreen:
		return vaultListView(m)

	case VaultDetailScreen:
		return vaultDetailView(m)

	case VaultTypeScreen:
		return vaultTypeView(m)

	case VaultCreateScreen:
		return vaultCreateView(m)
	}

	return tea.NewView("")
}

func vaultTypeView(m model) tea.View {
	s := "Выберите тип нового vault-элемента\n\n"

	for i, label := range itemTypeLabels {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		s += fmt.Sprintf("%s %s\n", cursor, label)
	}

	s += "\n↑↓ navigate | enter select | esc back"

	return tea.NewView(s)
}

func vaultCreateView(m model) tea.View {
	s := fmt.Sprintf("Create vault (%s)\n\n", m.selectedType)

	fields := createFieldsFor(&m)
	labels := createLabelsFor(m.selectedType)

	for i, f := range fields {
		s += labels[i] + ":\n"
		s += f.View()
		s += "\n\n"
	}

	s += "Tab - switch field"
	s += "\nEnter - create"
	s += "\nEsc - back"

	if m.vaultError != "" {
		s += "\nError: " + m.vaultError
	}

	return tea.NewView(s)
}

func menuView(m model) tea.View {
	s := "Password Manager\n\n"

	for i, item := range m.items {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		s += fmt.Sprintf("%s %s\n", cursor, item)
	}

	s += "\n↑↓ navigate | enter select"

	return tea.NewView(s)
}

func loginView(m model) tea.View {
	s := "Login\n\n"

	s += "Username:\n"
	s += m.username.View()
	s += "\n\n"

	s += "Password:\n"
	s += m.password.View()

	s += "\n\nTab - switch field"
	s += "\nEnter - login"
	s += "\nEsc - back"

	if m.registerOK != "" {
		s += "\n" + m.registerOK
	}

	if m.loginError != "" {
		s += "\nError: " + m.loginError
	}

	return tea.NewView(s)
}

func registerView(m model) tea.View {
	s := "Register\n\n"

	s += "Username:\n"
	s += m.regUsername.View()
	s += "\n\n"

	s += "Password:\n"
	s += m.regPassword.View()

	s += "\n\nTab - switch field"
	s += "\nEnter - register"
	s += "\nEsc - back"

	if m.registerError != "" {
		s += "\nError: " + m.registerError
	}

	return tea.NewView(s)
}

func vaultView(m model) tea.View {
	s := "Vault menu\n\n"

	items := []string{
		"List all vault",
		"Create new vault",
		"Exit",
	}

	for i, item := range items {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		s += fmt.Sprintf("%s %s\n", cursor, item)
	}

	if m.vaultError != "" {
		s += "\nError: " + m.vaultError
	}

	s += "\n\n↑↓ navigate | enter select"

	return tea.NewView(s)
}

func vaultDetailView(m model) tea.View {
	if m.vaultDetail == nil {
		return tea.NewView("Нет данных\n\nEsc - back")
	}

	item := m.vaultDetail

	s := "Vault item\n\n"
	s += fmt.Sprintf("ID:    %d\n", item.ID)
	s += fmt.Sprintf("Title: %s\n", item.Title)
	s += fmt.Sprintf("Type:  %s\n\n", item.Type)

	if secret, err := decodeSecretData(item.Type, item.Data); err != nil {
		s += fmt.Sprintf("Data: <ошибка: %s>\n", err)
	} else {
		s += fmt.Sprintf("Data: %s\n", secret)
	}

	s += fmt.Sprintf("Metadata: %s\n", formatMetadata(item.Metadata))

	if m.vaultStatus != "" {
		s += "\n" + m.vaultStatus
	}

	if m.vaultError != "" {
		s += "\nError: " + m.vaultError
	}

	s += "\n\nEsc - back"
	if item.Type == string(service_model.ItemBinary) {
		s += " | s - save file"
	}

	return tea.NewView(s)
}

func vaultListView(m model) tea.View {
	s := "Your vault\n\n"

	if len(m.vaultItems) == 0 {
		s += "Vault is empty"
	} else {
		for i, item := range m.vaultItems {
			cursor := "  "
			if m.cursor == i {
				cursor = "> "
			}

			s += fmt.Sprintf(
				"%s%d) %d | %s | %s\n",
				cursor,
				i+1,
				item.ID,
				item.Title,
				item.Type,
			)
		}

		s += "\n↑↓ navigate"
		s += " | enter - view details"
		s += " | s - save file (для BINARY)"
	}

	if m.vaultStatus != "" {
		s += "\n" + m.vaultStatus
	}

	if m.vaultError != "" {
		s += "\nError: " + m.vaultError
	}

	s += "\n\nEsc - back"

	return tea.NewView(s)
}
