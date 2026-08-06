package api

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"
)

type Client struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

func NewClient(baseURL string) *Client {
	caCert, err := os.ReadFile("server.crt")
	if err != nil {
		return nil
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caCert) {
		return nil
	}

	return &Client{
		BaseURL: baseURL,
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: pool,
				},
			},
		},
	}
}
