package main

import (
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
)

type StationResponse struct {
	Name        string  `json:"name"`
	Temperature string  `json:"weatherTemperature"`
	IconLink    string  `json:"weatherIconLink"`
	Proximity   float64 `json:"proximity"`
}

type Handler struct {
	transitAPIClient *TransitAPIClient
	weatherAPIClient *WeatherAPIClient
}

func (h *Handler) getStations(w http.ResponseWriter, r *http.Request) {
	in, ok := r.URL.Query()["in"]
	if !ok || len(in) < 1 {
		writeErrorResponse(w, http.StatusBadRequest, "Url Param 'in' is missing or invalid. Expected format: {lat},{lng}[;r={radius}]")
		return
	}
	var lat float64
	var lng float64
	var radius int = 1000 // random default value
	var err error
	coordinates := strings.Split(in[0], ",")
	lat, err = strconv.ParseFloat(coordinates[0], 64)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Invalid latitude format")
		return
	}
	lng, err = strconv.ParseFloat(coordinates[1], 64)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Invalid longitude format")
		return
	}
	rad, ok := r.URL.Query()["r"]
	if ok || len(rad) == 1 {
		radius, err = strconv.Atoi(rad[0])
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Invalid radius format")
			return
		}
	}
	resp, err := h.transitAPIClient.GetStations(r.Context(), lat, lng, radius)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	ret := make([]StationResponse, 0)
	var weatherCache sync.Map // If stations are in a same location doesn't make sense to ask for a weather for the same location
	// Cache is unique for a starting location. That why the cache isn't global
	wg := sync.WaitGroup{}
	wg.Add(len(resp))
	for _, place := range resp {
		go func(place Place) {
			defer wg.Done()
			station := StationResponse{
				Name: place.Name,
			}
			station.Proximity = distance(lat, lng, place.Location.Lat, place.Location.Lng)
			var weather *Observation
			distance := int(math.Floor(station.Proximity))
			weatherCacheItem, ok := weatherCache.Load(distance)
			// cache is empty
			if !ok {
				weather, _ = h.weatherAPIClient.GetWeather(r.Context(), place.Location.Lat, place.Location.Lng)
				weatherCache.Store(distance, weather)
			} else if weatherCacheItem != nil {
				weather = weatherCacheItem.(*Observation)
			}
			if weather == nil {
				weather = &Observation{}
			}
			station.IconLink = weather.IconLink
			station.Temperature = weather.Temperature
			ret = append(ret, station)
		}(place)
	}
	wg.Wait()
	writeJSONResponse(w, http.StatusOK, ret)
}

func (h *Handler) getSwagger(w http.ResponseWriter, r *http.Request) {
	workDir, _ := os.Getwd()
	fp := path.Join(workDir, "/swagger/swagger.yaml")
	http.ServeFile(w, r, fp)
}
