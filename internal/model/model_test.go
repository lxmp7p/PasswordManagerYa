package model

import (
	"testing"
	"time"
)

func TestItemTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		itemType ItemType
		expected string
	}{
		{"Login type", ItemLogin, "LOGIN_PASSWORD"},
		{"Text type", ItemText, "TEXT"},
		{"Binary type", ItemBinary, "BINARY"},
		{"Bank card type", ItemBankCard, "BANK_CARD"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.itemType) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(tt.itemType))
			}
		})
	}
}

func TestVaultItemTypes(t *testing.T) {
	tests := []struct {
		name     string
		item     VaultItem
		itemType ItemType
	}{
		{
			name: "Login item",
			item: VaultItem{
				ID:     1,
				UserID: 1,
				Type:   ItemLogin,
				Title:  "Gmail",
			},
			itemType: ItemLogin,
		},
		{
			name: "Text item",
			item: VaultItem{
				ID:     2,
				UserID: 1,
				Type:   ItemText,
				Title:  "Secret Note",
			},
			itemType: ItemText,
		},
		{
			name: "Binary item",
			item: VaultItem{
				ID:     3,
				UserID: 1,
				Type:   ItemBinary,
				Title:  "Private Key",
			},
			itemType: ItemBinary,
		},
		{
			name: "Bank card item",
			item: VaultItem{
				ID:     4,
				UserID: 1,
				Type:   ItemBankCard,
				Title:  "Visa",
			},
			itemType: ItemBankCard,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.item.Type != tt.itemType {
				t.Errorf("expected type %v, got %v", tt.itemType, tt.item.Type)
			}
		})
	}
}

func TestVaultItemTimestamps(t *testing.T) {
	now := time.Now()
	later := now.Add(1 * time.Hour)

	item := VaultItem{
		ID:        1,
		UserID:    1,
		CreatedAt: now,
		UpdatedAt: later,
	}

	if item.CreatedAt.After(item.UpdatedAt) {
		t.Error("CreatedAt should not be after UpdatedAt")
	}

	if !item.UpdatedAt.After(item.CreatedAt) {
		t.Error("UpdatedAt should be after CreatedAt")
	}
}

func TestVaultItemSecretData(t *testing.T) {
	tests := []struct {
		name       string
		secretData []byte
	}{
		{"Empty data", []byte("")},
		{"Small data", []byte("secret")},
		{"Large data", make([]byte, 10000)},
		{"Binary data", []byte{0x00, 0xFF, 0xAA, 0xBB}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := VaultItem{
				ID:         1,
				SecretData: tt.secretData,
			}

			if len(item.SecretData) != len(tt.secretData) {
				t.Errorf("expected length %d, got %d", len(tt.secretData), len(item.SecretData))
			}
		})
	}
}

func TestVaultItemZeroValue(t *testing.T) {
	var item VaultItem

	if item.ID != 0 {
		t.Errorf("expected ID 0, got %d", item.ID)
	}

	if item.UserID != 0 {
		t.Errorf("expected UserID 0, got %d", item.UserID)
	}

	if item.Type != "" {
		t.Errorf("expected empty Type, got %v", item.Type)
	}

	if item.Title != "" {
		t.Errorf("expected empty Title, got %q", item.Title)
	}

	if len(item.SecretData) != 0 {
		t.Errorf("expected empty SecretData, got %v", item.SecretData)
	}

	if !item.CreatedAt.IsZero() {
		t.Errorf("expected zero CreatedAt, got %v", item.CreatedAt)
	}

	if !item.UpdatedAt.IsZero() {
		t.Errorf("expected zero UpdatedAt, got %v", item.UpdatedAt)
	}
}

func TestMultipleVaultItems(t *testing.T) {
	now := time.Now()

	items := []VaultItem{
		{
			ID:         1,
			UserID:     100,
			Type:       ItemLogin,
			Title:      "Gmail",
			SecretData: []byte("gmail_secret"),
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         2,
			UserID:     100,
			Type:       ItemText,
			Title:      "Note",
			SecretData: []byte("note_content"),
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         3,
			UserID:     200,
			Type:       ItemBankCard,
			Title:      "Visa",
			SecretData: []byte("card_data"),
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	if len(items) != 3 {
		t.Errorf("expected 3 items, got %d", len(items))
	}

	user100Items := 0
	for _, item := range items {
		if item.UserID == 100 {
			user100Items++
		}
	}

	if user100Items != 2 {
		t.Errorf("expected 2 items for user 100, got %d", user100Items)
	}
}

func TestVaultItemTypeString(t *testing.T) {
	item := VaultItem{
		Type: ItemLogin,
	}

	typeStr := string(item.Type)
	if typeStr != "LOGIN_PASSWORD" {
		t.Errorf("expected 'LOGIN_PASSWORD', got %q", typeStr)
	}
}

func TestItemTypeComparison(t *testing.T) {
	type1 := ItemLogin
	type2 := ItemType("LOGIN_PASSWORD")

	if type1 != type2 {
		t.Error("ItemType comparison failed")
	}

	if type1 == ItemText {
		t.Error("different ItemTypes should not be equal")
	}
}
