package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"shab-checker/pkg/diff"
)

type Client struct {
	ServiceURL string
}

func NewClient(serviceURL string) *Client {
	return &Client{
		ServiceURL: serviceURL,
	}
}

func (c *Client) SendChanges(date string, changes []diff.ChangedCell) error {
	payload, err := json.Marshal(changes)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.ServiceURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send changes, status code: %d", resp.StatusCode)
	}

	return nil
}
