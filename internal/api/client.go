package api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type APIClient struct {
	Host     string
	APIKey   string
	Site     string
	Insecure bool
	client   *http.Client
}

func NewAPIClient(host, apiKey, site string, insecure bool) *APIClient {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	// Ensure host doesn't have trailing slash
	host = strings.TrimSuffix(host, "/")

	return &APIClient{
		Host:     host,
		APIKey:   apiKey,
		Site:     site,
		Insecure: insecure,
		client:   httpClient,
	}
}

func (c *APIClient) doRequest(method, path string) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.Host, path)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-KEY", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (c *APIClient) ListClients() ([]Client, error) {
	path := fmt.Sprintf("/proxy/network/api/s/%s/stat/sta", c.Site)

	body, err := c.doRequest("GET", path)
	if err != nil {
		return nil, err
	}

	var response ClientsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if response.Meta.RC != "ok" {
		return nil, fmt.Errorf("API returned error: %s", response.Meta.RC)
	}

	return response.Data, nil
}

func (c *APIClient) ListSites() ([]interface{}, error) {
	path := "/proxy/network/api/self/sites"

	body, err := c.doRequest("GET", path)
	if err != nil {
		return nil, err
	}

	var response APIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if response.Meta.RC != "ok" {
		return nil, fmt.Errorf("API returned error: %s", response.Meta.RC)
	}

	return response.Data, nil
}
