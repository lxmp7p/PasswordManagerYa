package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

const (
	POST = "POST"
)

func (c *Client) Login(login, password string) error {

	body, _ := json.Marshal(LoginRequest{
		Login:    login,
		Password: password,
	})

	req, err := http.NewRequest(
		POST,
		c.BaseURL+"/auth/login",
		bytes.NewReader(body),
	)

	if err != nil {
		return err
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	resp, err := c.Client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	var result LoginResponse

	err = json.NewDecoder(resp.Body).Decode(&result)

	if err != nil {
		return err
	}

	c.Token = result.Token

	fmt.Println(c)

	return nil
}
