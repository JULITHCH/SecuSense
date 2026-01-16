package synthesia

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/secusense/backend/config"
)

type Client struct {
	apiKey     string
	baseURL    string
	webhookURL string
	avatarID   string
	httpClient *http.Client
}

func NewClient(cfg config.SynthesiaConfig) *Client {
	return &Client{
		apiKey:     cfg.APIKey,
		baseURL:    cfg.BaseURL,
		webhookURL: cfg.WebhookURL,
		avatarID:   cfg.AvatarID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type createVideoRequest struct {
	Test        bool         `json:"test"`
	Title       string       `json:"title"`
	Description string       `json:"description,omitempty"`
	Input       []videoInput `json:"input"`
	CallbackID  string       `json:"callbackId,omitempty"`
	Visibility  string       `json:"visibility"`
	AspectRatio string       `json:"aspectRatio,omitempty"`
}

type videoInput struct {
	Avatar         string          `json:"avatar"`
	AvatarSettings *avatarSettings `json:"avatarSettings,omitempty"`
	Background     string          `json:"background"`
	ScriptText     string          `json:"scriptText"`
}

type avatarSettings struct {
	Voice string `json:"voice,omitempty"`
	Style string `json:"style"`
}

type createVideoResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type videoStatusResponse struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	Download string `json:"download,omitempty"`
}

func (c *Client) CreateVideo(ctx context.Context, script string, title string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("synthesia API key not configured")
	}

	reqBody := createVideoRequest{
		Test:        false,
		Title:       title,
		Visibility:  "private",
		AspectRatio: "16:9",
		Input: []videoInput{
			{
				Avatar:     c.avatarID,
				Background: "green_screen",
				ScriptText: script,
				AvatarSettings: &avatarSettings{
					Style: "rectangular",
				},
			},
		},
	}

	if c.webhookURL != "" {
		reqBody.CallbackID = c.webhookURL
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/videos", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("synthesia returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var videoResp createVideoResponse
	if err := json.NewDecoder(resp.Body).Decode(&videoResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return videoResp.ID, nil
}

func (c *Client) GetVideoStatus(ctx context.Context, videoID string) (string, string, error) {
	if c.apiKey == "" {
		return "", "", fmt.Errorf("synthesia API key not configured")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/videos/"+videoID, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("synthesia returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var statusResp videoStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		return "", "", fmt.Errorf("failed to decode response: %w", err)
	}

	return statusResp.Status, statusResp.Download, nil
}
