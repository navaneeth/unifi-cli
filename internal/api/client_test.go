package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewAPIClient(t *testing.T) {
	client := NewAPIClient("https://example.com", "test-key", "default", true)

	if client.Host != "https://example.com" {
		t.Errorf("Expected host 'https://example.com', got '%s'", client.Host)
	}
	if client.APIKey != "test-key" {
		t.Errorf("Expected APIKey 'test-key', got '%s'", client.APIKey)
	}
	if client.Site != "default" {
		t.Errorf("Expected site 'default', got '%s'", client.Site)
	}
	if client.Insecure != true {
		t.Errorf("Expected insecure 'true', got '%v'", client.Insecure)
	}
	if client.client == nil {
		t.Error("HTTP client should not be nil")
	}
}

func TestNewAPIClient_TrimTrailingSlash(t *testing.T) {
	client := NewAPIClient("https://example.com/", "test-key", "default", true)

	if client.Host != "https://example.com" {
		t.Errorf("Expected host without trailing slash, got '%s'", client.Host)
	}
}

func TestAPIClient_ListClients_Success(t *testing.T) {
	// Create mock response
	mockClients := []Client{
		{
			MAC:      "aa:bb:cc:dd:ee:ff",
			Name:     "TestDevice",
			IP:       "192.168.1.100",
			IsWired:  true,
			Hostname: "test-host",
		},
	}

	mockResponse := ClientsResponse{
		Meta: Meta{RC: "ok"},
		Data: mockClients,
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Verify API key header
		apiKey := r.Header.Get("X-API-KEY")
		if apiKey != "test-key" {
			t.Errorf("Expected API key 'test-key', got '%s'", apiKey)
		}

		// Verify path
		expectedPath := "/proxy/network/api/s/default/stat/sta"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path '%s', got '%s'", expectedPath, r.URL.Path)
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client and test
	client := NewAPIClient(server.URL, "test-key", "default", true)
	clients, err := client.ListClients()

	if err != nil {
		t.Fatalf("ListClients() returned error: %v", err)
	}

	if len(clients) != 1 {
		t.Fatalf("Expected 1 client, got %d", len(clients))
	}

	if clients[0].MAC != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("Expected MAC 'aa:bb:cc:dd:ee:ff', got '%s'", clients[0].MAC)
	}
}

func TestAPIClient_ListClients_APIError(t *testing.T) {
	mockResponse := ClientsResponse{
		Meta: Meta{RC: "error"},
		Data: []Client{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, "test-key", "default", true)
	_, err := client.ListClients()

	if err == nil {
		t.Error("Expected error for API error response")
	}
}

func TestAPIClient_ListClients_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, "test-key", "default", true)
	_, err := client.ListClients()

	if err == nil {
		t.Error("Expected error for HTTP 401 response")
	}
}

func TestAPIClient_ListClients_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, "test-key", "default", true)
	_, err := client.ListClients()

	if err == nil {
		t.Error("Expected error for invalid JSON response")
	}
}

func TestAPIClient_ListSites_Success(t *testing.T) {
	mockResponse := APIResponse{
		Meta: Meta{RC: "ok"},
		Data: []interface{}{
			map[string]interface{}{
				"name": "default",
				"desc": "Default Site",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify path
		expectedPath := "/proxy/network/api/self/sites"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path '%s', got '%s'", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, "test-key", "default", true)
	sites, err := client.ListSites()

	if err != nil {
		t.Fatalf("ListSites() returned error: %v", err)
	}

	if len(sites) != 1 {
		t.Fatalf("Expected 1 site, got %d", len(sites))
	}
}

func TestAPIClient_ListSites_APIError(t *testing.T) {
	mockResponse := APIResponse{
		Meta: Meta{RC: "error"},
		Data: []interface{}{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, "test-key", "default", true)
	_, err := client.ListSites()

	if err == nil {
		t.Error("Expected error for API error response")
	}
}

func TestAPIClient_doRequest_Success(t *testing.T) {
	expectedBody := `{"test":"data"}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("X-API-KEY") != "test-key" {
			t.Error("Missing or incorrect X-API-KEY header")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("Missing or incorrect Content-Type header")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedBody))
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, "test-key", "default", true)
	body, err := client.doRequest("GET", "/test")

	if err != nil {
		t.Fatalf("doRequest() returned error: %v", err)
	}

	if string(body) != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, string(body))
	}
}
