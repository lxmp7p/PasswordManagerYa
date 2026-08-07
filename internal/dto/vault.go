package dto

import (
	"encoding/json"
	"time"
)

type VaultCreateInDto struct {
	Type     string            `json:"type"`
	Title    string            `json:"title"`
	Data     json.RawMessage   `json:"data"`
	Metadata map[string]string `json:"metadata"`
}

type VaultRequestGet struct {
	VaultId int64 `json:"vault_id"`
}

type VaultResponseGet struct {
	ID        int64             `json:"id"`
	Type      string            `json:"type"`
	Title     string            `json:"title"`
	Data      any               `json:"data"`
	Metadata  map[string]string `json:"metadata"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}
