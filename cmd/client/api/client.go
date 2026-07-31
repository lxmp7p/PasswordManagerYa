package api

import (
	"crypto/tls"
	"net/http"
)

type Client struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // TODO: need to add cert
				},
			},
		},
	}
}
