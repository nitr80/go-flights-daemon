package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"nitr80.com/flights-daemon/internal/cache"
	"nitr80.com/flights-daemon/internal/handler"
	"nitr80.com/flights-daemon/internal/model"
	"nitr80.com/flights-daemon/internal/service"
)

func TestFlightsEndToEnd(t *testing.T) {
	cache := cache.NewFlightCacheImpl()

	err := cache.LoadFlights([]string{
		"../../internal/cache/testdata/test1.csv",
		"../../internal/cache/testdata/test2.csv",
	})
	if err != nil {
		t.Fatalf("failed to load flights: %v", err)
	}

	svc := &service.FlightServiceImpl{Cache: cache}
	flightHandler := &handler.FlightHandlerImpl{Service: svc}
	statusHandler := &handler.StatusHandlerImpl{Cache: cache}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /flights", flightHandler.GetFlights)
	mux.HandleFunc("GET /status", statusHandler.GetStatus)

	server := httptest.NewServer(mux)
	defer server.Close()

	response, err := http.Get(server.URL + "/flights?departure=CGN&destination=STR&departure_date=2022-06-10&return_date=2022-06-16")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	var body model.QueryResponseDTO
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.FlightOffers != 15 {
		t.Errorf("expected 15 flight offers, got %d", body.FlightOffers)
	}
	if len(body.Flights) != 15 {
		t.Errorf("expected 15 flights, got %d", len(body.Flights))
	}
	if len(body.Flights) > 0 {
		first := body.Flights[0]
		if first.Departure != "CGN" || first.Destination != "STR" {
			t.Errorf("unexpected route: %s -> %s", first.Departure, first.Destination)
		}
		if first.DepartureDate != "2022-06-10" || first.ReturnDate != "2022-06-16" {
			t.Errorf("unexpected dates: %s / %s", first.DepartureDate, first.ReturnDate)
		}
		if first.Carrier != "LH" {
			t.Errorf("expected carrier LH, got %s", first.Carrier)
		}
	}
}
