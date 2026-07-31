package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type VaultCreate struct {
	Type     string            `json:"type"`
	Title    string            `json:"title"`
	Data     any               `json:"data"`
	Metadata map[string]string `json:"metadata"`
}

func (c *Client) CreateVault(item VaultCreate) error {
	body, _ := json.Marshal(item)

	req, _ := http.NewRequest(
		"POST",
		c.BaseURL+"/vault",
		bytes.NewReader(body),
	)

	req.Header.Set(
		"Authorization",
		"Bearer "+c.Token,
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	resp, err := c.Client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}

type VaultItem struct {
	ID     int64  `json:"id"`
	Type   string `json:"type"`
	Title  string `json:"title"`
	Secret []byte `json:"secret_data"`
}

func (c *Client) ListVault() ([]VaultItem, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		c.BaseURL+"/vault",
		nil,
	)

	if err != nil {
		return nil, err
	}

	req.Header.Set(
		"Authorization",
		"Bearer "+c.Token,
	)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"vault list failed: %s",
			resp.Status,
		)
	}

	var items []VaultItem

	err = json.NewDecoder(resp.Body).Decode(&items)

	return items, err
}

func (c *Client) GetVaultByID(id int64) (VaultItem, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		c.BaseURL+"/vault/"+strconv.FormatInt(id, 10),
		nil,
	)

	if err != nil {
		return VaultItem{}, err
	}

	req.Header.Set(
		"Authorization",
		"Bearer "+c.Token,
	)

	resp, err := c.Client.Do(req)
	if err != nil {
		return VaultItem{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return VaultItem{}, fmt.Errorf(
			"vault list failed: %s",
			resp.Status,
		)
	}

	var item VaultItem

	err = json.NewDecoder(resp.Body).Decode(&item)

	return item, err
}

type VaultCreateRequest struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Data     string `json:"data"`
	Metadata string `json:"metadata"`
}
