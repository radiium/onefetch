package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ===============================
// Client Interface
// ===============================
type OneFichierClient interface {
	GetFileInfo(fileURL string) (*OneFichierInfoResponse, error)
	GetDownloadToken(fileURL string) (*OneFichierTokenResponse, error)
	DownloadFile(downloadURL string, offset int64) (io.ReadCloser, int64, int, error)
}

// ===============================
// Client Struct
// ===============================
type oneFichierClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// ===============================
// Data Structures
// ===============================

// OneFichierInfoResponse response of /file/info.cgi
type OneFichierInfoResponse struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	// Date        time.Time `json:"date"`
	Checksum    string  `json:"checksum"`
	ContentType string  `json:"content-type"`
	Description *string `json:"description,omitempty"`
	Pass        int     `json:"pass"`
	Path        string  `json:"path"`
	FolderID    string  `json:"folder_id"`
}

// OneFichierTokenResponse response of /download/get_token.cgi
type OneFichierTokenResponse struct {
	URL     string  `json:"url"`
	Status  string  `json:"status"`
	Message *string `json:"message,omitempty"`
}

// ===============================
// Client Constructor
// ===============================
func NewOneFichierClient(baseURL string, apiKey string) OneFichierClient {
	return &oneFichierClient{
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
// POST /file/info.cgi
// ===============================
func (c *oneFichierClient) GetFileInfo(fileURL string) (*OneFichierInfoResponse, error) {
	payload := map[string]string{"url": fileURL}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/file/info.cgi", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result OneFichierInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ===============================
// POST /download/get_token.cgi
// ===============================
func (c *oneFichierClient) GetDownloadToken(fileURL string) (*OneFichierTokenResponse, error) {
	payload := map[string]string{"url": fileURL}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/download/get_token.cgi", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result OneFichierTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Status != "OK" {
		return nil, fmt.Errorf("failed to get token: %v", result.Message)
	}

	return &result, nil
}

// ===============================
// GET download the file
// ===============================
func (c *oneFichierClient) DownloadFile(downloadURL string, offset int64) (io.ReadCloser, int64, int, error) {
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return nil, 0, 0, err
	}

	if offset > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", offset))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, 0, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		resp.Body.Close()
		return nil, 0, 0, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	return resp.Body, resp.ContentLength, resp.StatusCode, nil
}

func (c *oneFichierClient) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")
}
