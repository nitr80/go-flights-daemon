package service

import (
	"errors"
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

func TestGetFlightsSorting(t *testing.T) {
	newCache := func() *mockFlightCache {
		return &mockFlightCache{
			flights: []model.Flight{
				{Carrier: "LH", Price: 200},
				{Carrier: "BA", Price: 100},
				{Carrier: "AF", Price: 150},
			},
		}
	}

	t.Run("ascending by default", func(t *testing.T) {
		service := FlightServiceImpl{Cache: newCache()}
		result, err := service.GetFlights(model.QueryParameters{})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		assertPrices(t, result.Flights, []float64{100, 150, 200})
	})

	t.Run("ascending", func(t *testing.T) {
		service := FlightServiceImpl{Cache: newCache()}
		result, err := service.GetFlights(model.QueryParameters{SortDirection: "ASC"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		assertPrices(t, result.Flights, []float64{100, 150, 200})
	})

	t.Run("descending", func(t *testing.T) {
		service := FlightServiceImpl{Cache: newCache()}
		result, err := service.GetFlights(model.QueryParameters{SortDirection: "DESC"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		assertPrices(t, result.Flights, []float64{200, 150, 100})
	})
}

func assertPrices(t *testing.T, flights []model.Flight, expected []float64) {
	t.Helper()
	if len(flights) != len(expected) {
		t.Fatalf("expected %d flights, got %d", len(expected), len(flights))
	}
	for i, price := range expected {
		if flights[i].Price != price {
			t.Errorf("expected price %f at index %d, got %f", price, i, flights[i].Price)
		}
	}
}

func TestGetFlightsPagination(t *testing.T) {
	cache := &mockFlightCache{
		flights: []model.Flight{
			{Carrier: "LH", Price: 100},
			{Carrier: "LH", Price: 200},
			{Carrier: "LH", Price: 300},
			{Carrier: "LH", Price: 400},
			{Carrier: "LH", Price: 500},
		},
	}

	page, perPage := 2, 2
	service := FlightServiceImpl{Cache: cache}

	result, err := service.GetFlights(
		model.QueryParameters{CurrentPage: &page, PerPage: &perPage},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.PageCount != 3 {
		t.Errorf("expected page count 3, got %d", result.PageCount)
	}
	if result.CurrentPage != 2 {
		t.Errorf("expected current page 2, got %d", result.CurrentPage)
	}
	if result.PerPage != 2 {
		t.Errorf("expected per page 2, got %d", result.PerPage)
	}
	if result.FlightOffers != 5 {
		t.Errorf("expected 5 flight offers, got %d", result.FlightOffers)
	}
	assertPrices(t, result.Flights, []float64{300, 400})
}

func TestGetFlightsPageOutOfBounds(t *testing.T) {
	cache := &mockFlightCache{
		flights: []model.Flight{{Carrier: "LH", Price: 100}},
	}

	page, perPage := 5, 2
	service := FlightServiceImpl{Cache: cache}

	_, err := service.GetFlights(
		model.QueryParameters{CurrentPage: &page, PerPage: &perPage},
	)

	if !errors.Is(err, ErrPageOutOfBounds) {
		t.Errorf("expected ErrPageOutOfBounds, got %v", err)
	}
}
