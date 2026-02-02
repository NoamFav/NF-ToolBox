package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nf-software/nf-toolbox/pkg/license"
)

const (
	// DefaultBaseURL is the production API URL
	DefaultBaseURL = "https://api.nf-software.com/api/v1"
	// DevBaseURL for local development
	DevBaseURL = "http://localhost:8080/api/v1"
)

// Client is the API client for the license server
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new API client
func NewClient() *Client {
	return &Client{
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewDevClient creates a client pointing to localhost
func NewDevClient() *Client {
	c := NewClient()
	c.baseURL = DevBaseURL
	return c
}

// Auth types

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	} `json:"user"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Tool types

type Tool struct {
	Name         string `json:"name"`
	DisplayName  string `json:"display_name"`
	Description  string `json:"description"`
	PriceMonthly int    `json:"price_monthly"`
	PriceYearly  int    `json:"price_yearly"`
	Owned        bool   `json:"owned"`
}

type ToolsResponse struct {
	Tools []Tool `json:"tools"`
}

// Activation types

type ActivationRequest struct {
	MachineFingerprint string `json:"machine_fingerprint"`
	Nickname           string `json:"nickname"`
	Platform           string `json:"platform"`
	Arch               string `json:"arch"`
}

type ActivationResponse struct {
	License     *license.SignedLicense `json:"license"`
	DownloadURL string                 `json:"download_url,omitempty"`
}

// Renew types

type RenewRequest struct {
	MachineFingerprint string `json:"machine_fingerprint"`
}

type RenewResponse struct {
	Denied  bool                   `json:"denied"`
	Reason  string                 `json:"reason,omitempty"`
	License *license.SignedLicense `json:"license,omitempty"`
}

// Device types

type Device struct {
	ID                 string    `json:"id"`
	Nickname           string    `json:"nickname"`
	Platform           string    `json:"platform"`
	Arch               string    `json:"arch"`
	MachineFingerprint string    `json:"machine_fingerprint"`
	ActivatedAt        time.Time `json:"activated_at"`
	LastSeen           time.Time `json:"last_seen"`
}

type DevicesResponse struct {
	Devices    []Device `json:"devices"`
	MaxDevices int      `json:"max_devices"`
	Count      int      `json:"count"`
}

// Auth methods

func (c *Client) Login(email, password string) (*LoginResponse, error) {
	req := LoginRequest{Email: email, Password: password}
	var resp LoginResponse
	if err := c.post("/auth/login", "", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) Register(email, password string) (*LoginResponse, error) {
	req := RegisterRequest{Email: email, Password: password}
	var resp LoginResponse
	if err := c.post("/auth/register", "", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Tools methods

func (c *Client) ListTools(token string) ([]Tool, error) {
	var resp ToolsResponse
	if err := c.get("/tools", token, &resp); err != nil {
		return nil, err
	}
	return resp.Tools, nil
}

// Activation methods

func (c *Client) Activate(token string, req *ActivationRequest) (*ActivationResponse, error) {
	var resp ActivationResponse
	if err := c.post("/activate", token, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) Renew(token string, req *RenewRequest) (*RenewResponse, error) {
	var resp RenewResponse
	if err := c.post("/renew", token, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetLicense(token, fingerprint string) (*license.SignedLicense, error) {
	var resp license.SignedLicense
	url := fmt.Sprintf("/license?fingerprint=%s", fingerprint)
	if err := c.get(url, token, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Device methods

func (c *Client) ListDevices(token string) (*DevicesResponse, error) {
	var resp DevicesResponse
	if err := c.get("/devices", token, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) DeactivateDevice(token, deviceID string) error {
	url := fmt.Sprintf("/devices/%s", deviceID)
	return c.delete(url, token)
}

// HTTP helpers

func (c *Client) get(path, token string, result interface{}) error {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return err
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return c.do(req, result)
}

func (c *Client) post(path, token string, body, result interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.baseURL+path, bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return c.do(req, result)
}

func (c *Client) delete(path, token string) error {
	req, err := http.NewRequest("DELETE", c.baseURL+path, nil)
	if err != nil {
		return err
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return fmt.Errorf("%s", errResp.Error)
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	if result != nil && len(body) > 0 {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("decode response failed: %w", err)
		}
	}

	return nil
}
