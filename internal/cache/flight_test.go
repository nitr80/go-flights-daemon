package cache

import (
	"testing"
	"time"

	"nitr80.com/check24-go-daemon/internal/model"
)

func TestGetFlights(t *testing.T) {
	departureDate := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	returnDate := time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC)

	cache := FlightCacheImpl{
		Flights: []model.Flight{
			{
				Departure:     "BER",
				Destination:   "LHR",
				DepartureDate: departureDate,
				ReturnDate:    &returnDate,
				Carrier:       "LH",
				Price:         150.00,
			},
			{
				Departure:     "BER",
				Destination:   "LHR",
				DepartureDate: departureDate,
				ReturnDate:    nil,
				Carrier:       "BA",
				Price:         100.00,
			},
			{
				Departure:     "BER",
				Destination:   "CDG",
				DepartureDate: departureDate,
				ReturnDate:    &returnDate,
				Carrier:       "AF",
				Price:         120.00,
			},
		},
	}

	t.Run("filters by departure, destination and departure date", func(t *testing.T) {
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

	t.Run("filters by return date", func(t *testing.T) {
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
}
