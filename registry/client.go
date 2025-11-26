package registry

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Client struct {
	baseURL  string
	username string
	password string
	client   *http.Client
}

type Repository struct {
	Name string `json:"name"`
}

type RepositoriesResponse struct {
	Repositories []string `json:"repositories"`
}

type TagsResponse struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

type RepositoryInfo struct {
	Name string   `json:"name"`
	Tags []string `json:"tags,omitempty"`
}

func NewClient() (*Client, error) {
	baseURL := os.Getenv("REGISTRY_URL")
	if baseURL == "" {
		return nil, fmt.Errorf("REGISTRY_URL environment variable is required")
	}

	// Remove trailing slash if present
	baseURL = strings.TrimSuffix(baseURL, "/")

	username := os.Getenv("REGISTRY_USERNAME")
	password := os.Getenv("REGISTRY_PASSWORD")

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				// Allow insecure registries in development
				InsecureSkipVerify: username == "" && password == "",
			},
		},
	}

	return &Client{
		baseURL:  baseURL,
		username: username,
		password: password,
		client:   httpClient,
	}, nil
}

func (c *Client) doRequest(ctx context.Context, path string) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, path)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add basic auth if credentials are provided
	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("registry returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

func (c *Client) ListRepositories(ctx context.Context) ([]string, error) {
	body, err := c.doRequest(ctx, "/v2/_catalog")
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	var response RepositoriesResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse repositories response: %w", err)
	}

	return response.Repositories, nil
}

func (c *Client) ListTags(ctx context.Context, repository string) ([]string, error) {
	path := fmt.Sprintf("/v2/%s/tags/list", repository)
	body, err := c.doRequest(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags for %s: %w", repository, err)
	}

	var response TagsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse tags response: %w", err)
	}

	// Handle case where repository has no tags
	if response.Tags == nil {
		return []string{}, nil
	}

	return response.Tags, nil
}

func (c *Client) ListRepositoriesWithTags(ctx context.Context) ([]RepositoryInfo, error) {
	repos, err := c.ListRepositories(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]RepositoryInfo, 0, len(repos))
	for _, repo := range repos {
		tags, err := c.ListTags(ctx, repo)
		if err != nil {
			// Log error but continue with other repositories
			fmt.Printf("Warning: failed to fetch tags for %s: %v\n", repo, err)
			result = append(result, RepositoryInfo{
				Name: repo,
				Tags: []string{},
			})
			continue
		}

		result = append(result, RepositoryInfo{
			Name: repo,
			Tags: tags,
		})
	}

	return result, nil
}

func (c *Client) Close() error {
	// HTTP client doesn't need explicit closing
	return nil
}

// GetRegistryURL returns the base URL of the registry
func (c *Client) GetRegistryURL() string {
	return c.baseURL
}
