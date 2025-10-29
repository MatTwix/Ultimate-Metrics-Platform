package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type OpenWeatherClient struct {
	httpClient *http.Client
	apiKey     string
}

func NewOpenWeatherClient(apiKey string) *OpenWeatherClient {
	return &OpenWeatherClient{
		httpClient: &http.Client{},
		apiKey:     apiKey,
	}
}

type WeatherData struct {
	Main struct {
		Temp float64 `json:"temp"`
	} `json:"main"`
}

func (c *OpenWeatherClient) GetCurrentTemperature(ctx context.Context, city string) (*WeatherData, error) {
	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric", city, c.apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var data WeatherData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return &data, nil
}
