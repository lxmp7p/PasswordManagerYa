package dto

import "encoding/json"

type VaultCreateInDto struct {
	Type     string            `json:"type"`
	Title    string            `json:"title"`
	Data     json.RawMessage   `json:"data"`
	Metadata map[string]string `json:"metadata"`
}
