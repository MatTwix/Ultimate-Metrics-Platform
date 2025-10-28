package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type GithubClient struct {
	httpClient *http.Client
	token      string
}

func NewGithubClient(token string) *GithubClient {
	return &GithubClient{
		httpClient: &http.Client{},
		token:      token,
	}
}

type RepoInfo struct {
	StrangazersCount int `json:"strangazers_count"`
}

func (c *GithubClient) GetRepoInfo(ctx context.Context, repoName string) (*RepoInfo, error) {
	if !strings.Contains(repoName, "/") {
		return nil, fmt.Errorf("invalid repo name format, expected 'owner/repo'")
	}

	url := fmt.Sprintf("https://api.github.con/repos/%s", repoName)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var info RepoInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return &info, nil
}
