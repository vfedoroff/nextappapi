package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Observation struct {
	Temperature string `json:"temperature"`
	IconLink    string `json:"iconLink"`
}

type WheaterObservationResponse struct {
	Observations struct {
		Location []struct {
			Observation []Observation `json:"observation"`
		} `json:"location"`
	} `json:"observations"`
}

func NewWeatherAPIClient(httpClient *http.Client, apiURL, apiKey string) *WeatherAPIClient {
	return &WeatherAPIClient{
		httpClient: httpClient,
		apiKey:     apiKey,
		apiURL:     apiURL,
	}
}

type WeatherAPIClient struct {
	httpClient *http.Client
	apiKey     string
	apiURL     string
}

func (c *WeatherAPIClient) GetWeather(ctx context.Context, lat float64, lng float64) (*Observation, error) {
	req, _ := http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/weather/1.0/report.json?apiKey=%s&product=observation&latitude=%f&longitude=%f&oneobservation=true&metric=false",
			c.apiURL, c.apiKey, lat, lng),
		nil)
	req = req.WithContext(ctx)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var wheaterObservationResponse WheaterObservationResponse
	err = decoder.Decode(&wheaterObservationResponse)
	if err != nil {
		return nil, err
	}
	if len(wheaterObservationResponse.Observations.Location) > 0 &&
		len(wheaterObservationResponse.Observations.Location[0].Observation) > 0 {
		ret := wheaterObservationResponse.Observations.Location[0].Observation[0]
		return &ret, nil
	}
	return &Observation{}, nil
}
