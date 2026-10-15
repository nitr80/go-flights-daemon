package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"nitr80.com/flights-daemon/internal/model"
	"nitr80.com/flights-daemon/internal/service"
)

type mockFlightService struct {
	response model.QueryResponse
	err      error

	receivedParameters model.QueryParameters
}

func (m *mockFlightService) GetFlights(
	queryParameters model.QueryParameters,
) (model.QueryResponse, error) {
	m.receivedParameters = queryParameters

	return m.response, m.err
}

func TestGetFlights(t *testing.T) {
	departureDate := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	returnDate := time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC)

	service := &mockFlightService{
		response: model.QueryResponse{
			Runtime:      0.001,
			FlightOffers: 2,
			PageCount:    1,
			CurrentPage:  1,
			PerPage:      10,
			AirlineOffers: []model.AirlineOffer{
				{
					Code:     "LH",
					PriceMin: 100,
					PriceMax: 200,
				},
			},
			Flights: []model.Flight{
				{
					Departure:     "BER",
					Destination:   "LHR",
					DepartureDate: departureDate,
					ReturnDate:    &returnDate,
					Carrier:       "LH",
					Price:         100,
				},
			},
		},
	}

	handler := FlightHandlerImpl{
		Service: service,
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/flights?departure=BER&destination=LHR&departure_date=2026-10-20&return_date=2026-10-27&airlines[]=LH&page=1&per_page=10&sort_direction=ASC",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.GetFlights(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	if recorder.Header().Get("Content-Type") != "application/json" {
		t.Errorf(
			"expected Content-Type application/json, got %s",
			recorder.Header().Get("Content-Type"),
		)
	}

	if service.receivedParameters.Departure != "BER" {
		t.Errorf(
			"expected departure BER, got %s",
			service.receivedParameters.Departure,
		)
	}

	if service.receivedParameters.Destination != "LHR" {
		t.Errorf(
			"expected destination LHR, got %s",
			service.receivedParameters.Destination,
		)
	}

	if !service.receivedParameters.DepartureDate.Equal(departureDate) {
		t.Errorf("unexpected departure date")
	}

	if service.receivedParameters.ReturnDate == nil ||
		!service.receivedParameters.ReturnDate.Equal(returnDate) {
		t.Errorf("unexpected return date")
	}

	if len(service.receivedParameters.Carrier) != 1 ||
		service.receivedParameters.Carrier[0] != "LH" {
		t.Errorf("unexpected airlines: %v", service.receivedParameters.Carrier)
	}

	if service.receivedParameters.CurrentPage == nil ||
		*service.receivedParameters.CurrentPage != 1 {
		t.Errorf("expected page 1")
	}

	if service.receivedParameters.PerPage == nil ||
		*service.receivedParameters.PerPage != 10 {
		t.Errorf("expected per_page 10")
	}

	if service.receivedParameters.SortDirection != "ASC" {
		t.Errorf(
			"expected sort direction ASC, got %s",
			service.receivedParameters.SortDirection,
		)
	}
}

func TestGetFlightsValidationErrors(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{"missing required params", "/flights?destination=LHR&departure_date=2026-10-20"},
		{"invalid departure date", "/flights?departure=BER&destination=LHR&departure_date=not-a-date"},
		{"invalid return date", "/flights?departure=BER&destination=LHR&departure_date=2026-10-20&return_date=not-a-date"},
		{"invalid page", "/flights?departure=BER&destination=LHR&departure_date=2026-10-20&page=abc"},
		{"zero page", "/flights?departure=BER&destination=LHR&departure_date=2026-10-20&page=0"},
		{"invalid per_page", "/flights?departure=BER&destination=LHR&departure_date=2026-10-20&per_page=abc"},
		{"zero per_page", "/flights?departure=BER&destination=LHR&departure_date=2026-10-20&per_page=0"},
		{"invalid sort direction", "/flights?departure=BER&destination=LHR&departure_date=2026-10-20&sort_direction=SIDEWAYS"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := FlightHandlerImpl{Service: &mockFlightService{}}
			recorder := httptest.NewRecorder()

			handler.GetFlights(recorder, httptest.NewRequest(http.MethodGet, tt.query, nil))

			if recorder.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", recorder.Code)
			}
		})
	}
}

func TestGetFlightsPageOutOfBounds(t *testing.T) {
	handler := FlightHandlerImpl{Service: &mockFlightService{err: service.ErrPageOutOfBounds}}
	recorder := httptest.NewRecorder()

	handler.GetFlights(
		recorder,
		httptest.NewRequest(http.MethodGet, "/flights?departure=BER&destination=LHR&departure_date=2026-10-20", nil),
	)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", recorder.Code)
	}
}

func TestGetFlightsResponseIncludesReturnDate(t *testing.T) {
	departureDate := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	returnDate := time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC)

	service := &mockFlightService{
		response: model.QueryResponse{
			Flights: []model.Flight{
				{
					Departure:     "BER",
					Destination:   "LHR",
					DepartureDate: departureDate,
					ReturnDate:    &returnDate,
					Carrier:       "LH",
					Price:         100,
				},
			},
		},
	}

	handler := FlightHandlerImpl{Service: service}
	recorder := httptest.NewRecorder()

	handler.GetFlights(
		recorder,
		httptest.NewRequest(http.MethodGet, "/flights?departure=BER&destination=LHR&departure_date=2026-10-20&return_date=2026-10-27", nil),
	)

	var body model.QueryResponseDTO
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(body.Flights) != 1 {
		t.Fatalf("expected 1 flight, got %d", len(body.Flights))
	}

	if body.Flights[0].ReturnDate != "2026-10-27" {
		t.Errorf("expected return date 2026-10-27, got %s", body.Flights[0].ReturnDate)
	}
}
