package model

import "time"

type ItemType string

const (
	ItemLogin    ItemType = "LOGIN_PASSWORD"
	ItemText     ItemType = "TEXT"
	ItemBinary   ItemType = "BINARY"
	ItemBankCard ItemType = "BANK_CARD"
)

type VaultItem struct {
	ID        int64
	UserID    int64
	Type      ItemType
	Title     string
	Secret    []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}
