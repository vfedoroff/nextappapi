package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Place struct {
	Name     string `json:"name"`
	Location struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"location"`
}

type StationsResponse struct {
	Stations []struct {
		Place `json:"place"`
	} `json:"stations"`
}

func NewTransitAPIClient(httpClient *http.Client, apiURL, apiKey string) *TransitAPIClient {
	return &TransitAPIClient{
		httpClient: httpClient,
		apiKey:     apiKey,
		apiURL:     apiURL,
	}
}

type TransitAPIClient struct {
	httpClient *http.Client
	apiURL     string
	apiKey     string
}

func (c *TransitAPIClient) GetStations(ctx context.Context, lat float64, lng float64, r int) ([]Place, error) {
	req, _ := http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/v8/stations?in=%f,%f;r=%d&apiKey=%s",
			c.apiURL, lat, lng, r, c.apiKey),
		nil)
	req = req.WithContext(ctx)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var stationsResponse StationsResponse
	err = decoder.Decode(&stationsResponse)
	var stations = make([]Place, 0)
	for _, station := range stationsResponse.Stations {
		stations = append(stations, station.Place)
	}
	return stations, err
}
