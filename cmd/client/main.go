package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"passwordmanager/cmd/client/api"
	service_model "passwordmanager/internal/model"
	"path/filepath"
	"strconv"

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
	VaultScreen
	VaultListScreen
	VaultIDScreen
	VaultTypeScreen   // выбор типа нового vault-элемента
	VaultCreateScreen // форма ввода данных под выбранный тип
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

	vaultItems []api.VaultItem
	vaultError string

	vaultID textinput.Model

	// Выбранный тип создаваемого vault-элемента.
	selectedType service_model.ItemType

	// Поле "Title" — общее для всех типов.
	createTitle textinput.Model
	// Индекс текущего сфокусированного поля в форме создания (0 = createTitle).
	createFocus int

	// Поля под конкретные типы. Показываются/используются только те,
	// что относятся к m.selectedType (см. createFieldsFor / createLabelsFor).
	vaultLogin    textinput.Model
	vaultPassword textinput.Model
	vaultText     textinput.Model
	vaultBinary   textinput.Model
	vaultNumber   textinput.Model
	vaultHolder   textinput.Model
	vaultMonth    textinput.Model
	vaultYear     textinput.Model
	vaultCVV      textinput.Model
}

type LoginResultMsg struct {
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

func initialModel(client *api.Client) model {
	username := textinput.New()
	username.Placeholder = "username"

	password := textinput.New()
	password.Placeholder = "password"
	password.EchoMode = textinput.EchoPassword

	vaultID := textinput.New()
	vaultID.Placeholder = "vault id"

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

	return model{
		screen: MenuScreen,
		client: client,

		items: []string{
			"Login",
			"Exit",
		},

		username:      username,
		password:      password,
		vaultID:       vaultID,
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

		m.vaultItems = []api.VaultItem{msg.item}
		m.screen = VaultListScreen

		return m, nil

	case VaultCreateResultMsg:
		if msg.err != nil {
			m.vaultError = msg.err.Error()
			return m, nil
		}

		m.screen = VaultScreen

		return m, loadVaultsCmd(m.client)
	}

	switch m.screen {

	case MenuScreen:
		return menuUpdate(m, msg)

	case LoginScreen:
		return loginUpdate(m, msg)

	case VaultScreen:
		return vaultUpdate(m, msg)

	case VaultListScreen:
		return vaultListUpdate(m, msg)

	case VaultIDScreen:
		return vaultIDUpdate(m, msg)

	case VaultTypeScreen:
		return vaultTypeUpdate(m, msg)

	case VaultCreateScreen:
		return vaultCreateUpdate(m, msg)
	}

	return m, nil
}

// ---------- helpers для динамической формы создания vault-элемента ----------

// createFieldsFor возвращает указатели на поля ввода, актуальные для
// текущего m.selectedType. Индекс 0 — всегда createTitle.
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

	return fields
}

// createLabelsFor — подписи полей в том же порядке, что и createFieldsFor.
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

	return labels
}

// resetCreateForm очищает и сбрасывает фокус всех полей формы создания.
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

	m.vaultLogin.Blur()
	m.vaultPassword.Blur()
	m.vaultText.Blur()
	m.vaultBinary.Blur()
	m.vaultNumber.Blur()
	m.vaultHolder.Blur()
	m.vaultMonth.Blur()
	m.vaultYear.Blur()
	m.vaultCVV.Blur()

	m.createFocus = 0
	m.createTitle.Focus()
}

// buildVaultData собирает map[string]any под конкретный тип из значений полей.
func buildVaultData(m model) map[string]any {
	data := map[string]any{}

	switch m.selectedType {
	case service_model.ItemLogin:
		data["login"] = m.vaultLogin.Value()
		data["password"] = m.vaultPassword.Value()
	case service_model.ItemText:
		data["text"] = m.vaultText.Value()
	// ItemBinary сюда не попадает: файл читается и кодируется отдельно,
	// см. createBinaryVaultCmd.
	case service_model.ItemBankCard:
		data["number"] = m.vaultNumber.Value()
		data["holder"] = m.vaultHolder.Value()
		data["month"] = m.vaultMonth.Value()
		data["year"] = m.vaultYear.Value()
		data["cvv"] = m.vaultCVV.Value()
	}

	return data
}

// createBinaryVaultCmd читает файл по указанному пути, кодирует его
// содержимое в base64 и отправляет вместе с именем файла. Формат
// {"name": ..., "data": <base64>} совпадает с FileData на сервере
// (json.Unmarshal сам раскодирует base64-строку в []byte).
func createBinaryVaultCmd(client *api.Client, title, path string) tea.Cmd {
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
			Metadata: map[string]string{},
		}

		err = client.CreateVault(item)
		return VaultCreateResultMsg{err: err}
	}
}

// ---------- screen updates ----------

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
			if m.selectedType == service_model.ItemBinary {
				return m, createBinaryVaultCmd(m.client, m.createTitle.Value(), m.vaultBinary.Value())
			}

			item := api.VaultCreate{
				Title:    m.createTitle.Value(),
				Type:     string(m.selectedType),
				Data:     buildVaultData(m),
				Metadata: map[string]string{},
			}

			return m, func() tea.Msg {
				err := m.client.CreateVault(item)
				return VaultCreateResultMsg{err: err}
			}
		}
	}

	return m, cmd
}

func vaultIDUpdate(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.vaultID, cmd = m.vaultID.Update(msg)

	switch msg := msg.(type) {

	case tea.KeyPressMsg:
		switch msg.String() {

		case "esc":
			m.screen = VaultScreen

		case "enter":
			id, err := strconv.ParseInt(m.vaultID.Value(), 10, 64)
			if err != nil {
				m.vaultError = "invalid vault id"
				return m, nil
			}

			return m, loadVaultById(m.client, id)
		}
	}

	return m, cmd
}

// menuUpdate реалезует логику для главного меню
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

			case 1:
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
				m.screen = VaultScreen
				return m, loadVaultsCmd(m.client)

			case 1:
				m.screen = VaultIDScreen
				m.vaultID.Focus()
				return m, nil

			case 2:
				m.screen = VaultTypeScreen
				m.cursor = 0
				return m, nil

			case 3:
				m.screen = MenuScreen
			}

		case "down":
			if m.cursor < 3 {
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
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {

		case "esc":
			m.screen = VaultScreen
			m.vaultItems = nil
			m.cursor = 0
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

// ---------- views ----------

func (m model) View() tea.View {
	switch m.screen {

	case MenuScreen:
		return menuView(m)

	case LoginScreen:
		return loginView(m)

	case VaultScreen:
		return vaultView(m)

	case VaultListScreen:
		return vaultListView(m)

	case VaultIDScreen:
		return vaultIDView(m)

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

func vaultIDView(m model) tea.View {
	s := "Enter vault ID\n\n"
	s += m.vaultID.View()
	s += "\n\nEnter - search"
	s += "\nEsc - back"

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

	if m.loginError != "" {
		s += "\nError: " + m.loginError
	}

	return tea.NewView(s)
}

func vaultView(m model) tea.View {
	s := "Vault menu\n\n"

	if len(m.vaultItems) > 0 {
		for _, item := range m.vaultItems {
			s += fmt.Sprintf(
				"%d | %s | %s\n",
				item.ID,
				item.Title,
				item.Type,
			)
		}

		s += "\nEsc - back"

		return tea.NewView(s)
	}

	items := []string{
		"List all vault",
		"Get vault by id",
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

func vaultListView(m model) tea.View {
	s := "Your vault\n\n"

	if len(m.vaultItems) == 0 {
		s += "Vault is empty"
	} else {
		for _, item := range m.vaultItems {
			s += fmt.Sprintf(
				"%d | %s | %s\n",
				item.ID,
				item.Title,
				item.Type,
			)
		}
	}

	s += "\n\nEsc - back"

	return tea.NewView(s)
}

func main() {
	client := api.NewClient("https://localhost:4443")
	p := tea.NewProgram(
		initialModel(client),
	)
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
