package unsplash

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/secusense/backend/config"
)

type Client struct {
	accessKey  string
	baseURL    string
	httpClient *http.Client
}

// Photo represents a photo from the Unsplash API
type Photo struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	AltDesc     string `json:"alt_description"`
	URLs        struct {
		Raw     string `json:"raw"`
		Full    string `json:"full"`
		Regular string `json:"regular"`
		Small   string `json:"small"`
		Thumb   string `json:"thumb"`
	} `json:"urls"`
	User struct {
		Name string `json:"name"`
	} `json:"user"`
}

// SearchResult represents the response from Unsplash search API
type SearchResult struct {
	Total      int     `json:"total"`
	TotalPages int     `json:"total_pages"`
	Results    []Photo `json:"results"`
}

func NewClient(cfg config.UnsplashConfig) *Client {
	if cfg.AccessKey == "" {
		log.Printf("Unsplash client initialized without API key - image search will be disabled")
		return &Client{
			accessKey: "",
			baseURL:   cfg.BaseURL,
			httpClient: &http.Client{
				Timeout: 10 * time.Second,
			},
		}
	}

	log.Printf("Unsplash client initialized: baseURL=%s", cfg.BaseURL)
	return &Client{
		accessKey: cfg.AccessKey,
		baseURL:   cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// IsAvailable returns true if the Unsplash client is configured
func (c *Client) IsAvailable() bool {
	return c.accessKey != ""
}

// SearchPhotos searches for photos matching the given keywords
func (c *Client) SearchPhotos(ctx context.Context, keywords string, perPage int) ([]Photo, error) {
	if !c.IsAvailable() {
		return nil, nil
	}

	if perPage <= 0 {
		perPage = 1
	}
	if perPage > 30 {
		perPage = 30
	}

	// Build the URL
	endpoint := fmt.Sprintf("%s/search/photos", c.baseURL)
	params := url.Values{}
	params.Set("query", keywords)
	params.Set("per_page", fmt.Sprintf("%d", perPage))
	params.Set("orientation", "landscape")

	reqURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Client-ID "+c.accessKey)
	req.Header.Set("Accept-Version", "v1")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unsplash returned status %d: %s", resp.StatusCode, string(body))
	}

	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Results, nil
}

// GetPhotoForKeywords searches for a photo and returns the first result
func (c *Client) GetPhotoForKeywords(ctx context.Context, keywords string) (*Photo, error) {
	photos, err := c.SearchPhotos(ctx, keywords, 1)
	if err != nil {
		return nil, err
	}

	if len(photos) == 0 {
		return nil, nil
	}

	return &photos[0], nil
}

// GetImageURL returns a suitable image URL (regular size for good quality)
func (p *Photo) GetImageURL() string {
	return p.URLs.Regular
}

// GetAltText returns the alt description for the photo
func (p *Photo) GetAltText() string {
	if p.AltDesc != "" {
		return p.AltDesc
	}
	if p.Description != "" {
		return p.Description
	}
	return "Stock photo"
}
