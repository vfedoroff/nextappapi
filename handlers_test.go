package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetStationsRemoteAPIReturnsError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()
	// Point clients to mock service
	transitAPIClient := NewTransitAPIClient(http.DefaultClient,
		ts.URL,
		"token")
	weatherAPIClient := NewWeatherAPIClient(http.DefaultClient,
		ts.URL,
		"token")
	handlers := Handler{transitAPIClient: transitAPIClient, weatherAPIClient: weatherAPIClient}
	req := httptest.NewRequest("GET", "http://localhost:8080/api/v1/stations?in=52.5251,13.3694;r=500", nil)
	w := httptest.NewRecorder()
	handlers.getStations(w, req)
	if w.Code != http.StatusInternalServerError {
		log.Fatalf("expected %d, actual: %d", http.StatusInternalServerError, w.Code)
	}
}
