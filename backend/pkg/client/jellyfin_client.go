package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ===============================
// Client Interface
// ===============================
type JellyfinClient interface {
	GetVirtualFolders(ctx context.Context) ([]VirtualFolder, error)
	RefreshLibrary(ctx context.Context) error
}

// ===============================
// Client Struct
// ===============================
type jellyfinClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// ===============================
// Data Structures
// ===============================

// Response of /Library/VirtualFolders
type VirtualFolder struct {
	Name           string   `json:"Name"`
	ItemID         string   `json:"ItemId"`
	Locations      []string `json:"Locations"`
	CollectionType string   `json:"CollectionType"`
}

// ===============================
// Client Constructor
// ===============================
func NewJellyfinClient(baseURL string, apiKey string) JellyfinClient {
	return &jellyfinClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 0, // Pas de timeout pour le body streaming
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				MaxConnsPerHost:     100,
			},
		},
	}
}

// ===============================
// GET /Library/VirtualFolders
// ===============================
func (c *jellyfinClient) GetVirtualFolders(ctx context.Context) ([]VirtualFolder, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/Library/VirtualFolders", nil)
	if err != nil {
		return nil, err
	}

	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error %d: %s", resp.StatusCode, string(data))
	}

	var folders []VirtualFolder
	if err := json.NewDecoder(resp.Body).Decode(&folders); err != nil {
		return nil, err
	}

	return folders, nil
}

// ===============================
// POST /Library/Refresh
// ===============================
func (c *jellyfinClient) RefreshLibrary(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/Library/Refresh", bytes.NewBuffer([]byte{}))
	if err != nil {
		return err
	}

	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error %d: %s", resp.StatusCode, string(data))
	}

	return nil
}

// ===============================
// POST /Items/:itemId/Refresh
// ===============================
func (c *jellyfinClient) RefreshItem(ctx context.Context, itemId string) error {
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/Items/"+itemId+"/Refresh", nil)
	if err != nil {
		return err
	}

	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error %d: %s", resp.StatusCode, string(data))
	}

	return nil
}

func (c *jellyfinClient) setHeaders(req *http.Request) {
	req.Header.Set("X-Emby-Token", c.apiKey)
}
