package cache

import (
	"testing"
	"time"

	"nitr80.com/flights-daemon/internal/model"
)

func TestGetFlights(t *testing.T) {
	departureDate := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	returnDate := time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC)

	berLhrReturn := model.Flight{
		Departure:     "BER",
		Destination:   "LHR",
		DepartureDate: departureDate,
		ReturnDate:    &returnDate,
		Carrier:       "LH",
		Price:         150.00,
	}
	berLhrOneWay := model.Flight{
		Departure:     "BER",
		Destination:   "LHR",
		DepartureDate: departureDate,
		ReturnDate:    nil,
		Carrier:       "BA",
		Price:         100.00,
	}
	berCdgReturn := model.Flight{
		Departure:     "BER",
		Destination:   "CDG",
		DepartureDate: departureDate,
		ReturnDate:    &returnDate,
		Carrier:       "AF",
		Price:         120.00,
	}
	cache := FlightCacheImpl{
		Flights: make(map[string][]model.Flight),
	}
	// Populate the cache using the same key-generation logic as production.
	for _, flight := range []model.Flight{
		berLhrReturn,
		berLhrOneWay,
		berCdgReturn,
	} {
		key := createKeyWithFlight(flight)
		cache.Flights[key] = append(cache.Flights[key], flight)
	}
	t.Run("returns flights matching departure, destination and departure date", func(t *testing.T) {
		filter := model.FilterParameters{
			Departure:     "BER",
			Destination:   "LHR",
			DepartureDate: departureDate,
		}
		flights := cache.GetFlights(filter)
		if len(flights) != 1 {
			t.Fatalf("expected 1 flight, got %d", len(flights))
		}
		if flights[0].Carrier != "BA" {
			t.Errorf("expected carrier BA, got %s", flights[0].Carrier)
		}
	})
	t.Run("returns flights matching return date", func(t *testing.T) {
		filter := model.FilterParameters{
			Departure:     "BER",
			Destination:   "LHR",
			DepartureDate: departureDate,
			ReturnDate:    &returnDate,
		}
		flights := cache.GetFlights(filter)
		if len(flights) != 1 {
			t.Fatalf("expected 1 flight, got %d", len(flights))
		}
		if flights[0].Carrier != "LH" {
			t.Errorf("expected carrier LH, got %s", flights[0].Carrier)
		}
	})
	t.Run("returns no flights for non-matching destination", func(t *testing.T) {
		filter := model.FilterParameters{
			Departure:     "BER",
			Destination:   "MUC",
			DepartureDate: departureDate,
		}
		flights := cache.GetFlights(filter)
		if len(flights) != 0 {
			t.Fatalf("expected 0 flights, got %d", len(flights))
		}
	})
	t.Run("returns all flights matching the same key", func(t *testing.T) {
		anotherFlight := berLhrOneWay
		anotherFlight.Carrier = "EW"
		anotherFlight.Price = 130.00
		key := createKeyWithFlight(anotherFlight)
		cache.Flights[key] = append(cache.Flights[key], anotherFlight)
		filter := model.FilterParameters{
			Departure:     "BER",
			Destination:   "LHR",
			DepartureDate: departureDate,
		}
		flights := cache.GetFlights(filter)
		if len(flights) != 2 {
			t.Fatalf("expected 2 flights, got %d", len(flights))
		}
	})
	t.Run("returns no flights for an unknown key", func(t *testing.T) {
		filter := model.FilterParameters{
			Departure:     "BER",
			Destination:   "MUC",
			DepartureDate: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
		}
		flights := cache.GetFlights(filter)
		if len(flights) != 0 {
			t.Fatalf("expected 0 flights, got %d", len(flights))
		}
	})

}

func TestLoadFlights(t *testing.T) {
	cache := NewFlightCacheImpl()

	if loading, err := cache.GetLoadStatus(); !loading || err != nil {
		t.Fatalf("expected loading before load, got loading=%v err=%v", loading, err)
	}

	err := cache.LoadFlights([]string{"testdata/test1.csv", "testdata/test2.csv"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if loading, err := cache.GetLoadStatus(); loading || err != nil {
		t.Errorf("expected finished without error, got loading=%v err=%v", loading, err)
	}

	departureDate := time.Date(2022, 6, 10, 0, 0, 0, 0, time.UTC)
	returnDate := time.Date(2022, 6, 16, 0, 0, 0, 0, time.UTC)

	filter := model.FilterParameters{
		Departure:     "CGN",
		Destination:   "STR",
		DepartureDate: departureDate,
		ReturnDate:    &returnDate,
	}

	flights := cache.GetFlights(filter)
	if len(flights) != 15 {
		t.Fatalf("expected 15 flights, got %d", len(flights))
	}

	for _, flight := range flights {
		if flight.Carrier != "LH" {
			t.Errorf("expected carrier LH, got %s", flight.Carrier)
		}
	}

	filter.Carrier = []string{"XX"}
	if filtered := cache.GetFlights(filter); len(filtered) != 0 {
		t.Errorf("expected 0 flights for unknown carrier, got %d", len(filtered))
	}
}

func TestLoadFlightsReportsFailure(t *testing.T) {
	cache := NewFlightCacheImpl()

	if err := cache.LoadFlights([]string{"testdata/does-not-exist.csv"}); err == nil {
		t.Fatal("expected an error for a missing file")
	}

	loading, err := cache.GetLoadStatus()
	if loading || err == nil {
		t.Errorf("expected finished with error, got loading=%v err=%v", loading, err)
	}
}
