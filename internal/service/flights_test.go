package service

import (
	"testing"
	"time"

	"nitr80.com/flights-daemon/internal/model"
)

type mockFlightCache struct {
	flights []model.Flight
}

func (m *mockFlightCache) GetFlights(filterParameters model.FilterParameters) []model.Flight {
	return m.flights
}

func TestGetFlights(t *testing.T) {
	departureDate := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	returnDate := time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC)

	flights := []model.Flight{
		{
			Departure:     "BER",
			Destination:   "LHR",
			DepartureDate: departureDate,
			ReturnDate:    &returnDate,
			Carrier:       "LH",
			Price:         100,
		},
		{
			Departure:     "BER",
			Destination:   "LHR",
			DepartureDate: departureDate,
			ReturnDate:    &returnDate,
			Carrier:       "LH",
			Price:         200,
		},
		{
			Departure:     "BER",
			Destination:   "LHR",
			DepartureDate: departureDate,
			ReturnDate:    &returnDate,
			Carrier:       "BA",
			Price:         150,
		},
	}

	cache := &mockFlightCache{
		flights: flights,
	}

	service := FlightServiceImpl{
		Cache: cache,
	}

	query := model.QueryParameters{
		Departure:     "BER",
		Destination:   "LHR",
		DepartureDate: departureDate,
		ReturnDate:    &returnDate,
	}

	result, err := service.GetFlights(query)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.FlightOffers != 3 {
		t.Errorf("expected 3 flight offers, got %d", result.FlightOffers)
	}

	if len(result.Flights) != 3 {
		t.Errorf("expected 3 flights, got %d", len(result.Flights))
	}

	if len(result.AirlineOffers) != 2 {
		t.Fatalf("expected 2 airline offers, got %d", len(result.AirlineOffers))
	}

	lh := findAirlineOffer(result.AirlineOffers, "LH")

	if lh == nil {
		t.Fatal("expected LH airline offer")
	}

	if lh.PriceMin != 100 {
		t.Errorf("expected LH minimum price 100, got %f", lh.PriceMin)
	}

	if lh.PriceMax != 200 {
		t.Errorf("expected LH maximum price 200, got %f", lh.PriceMax)
	}

	ba := findAirlineOffer(result.AirlineOffers, "BA")

	if ba == nil {
		t.Fatal("expected BA airline offer")
	}

	if ba.PriceMin != 150 || ba.PriceMax != 150 {
		t.Errorf(
			"expected BA prices 150/150, got %f/%f",
			ba.PriceMin,
			ba.PriceMax,
		)
	}

	if result.Runtime < 0 {
		t.Errorf("expected non-negative runtime, got %f", result.Runtime)
	}
}

func findAirlineOffer(
	offers []model.AirlineOffer,
	code string,
) *model.AirlineOffer {
	for i := range offers {
		if offers[i].Code == code {
			return &offers[i]
		}
	}

	return nil
}
