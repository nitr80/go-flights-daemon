package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"nitr80.com/flights-daemon/internal/model"
	"nitr80.com/flights-daemon/internal/service"
)

type FlightService interface {
	GetFlights(queryParameters model.QueryParameters) (model.QueryResponse, error)
}

type FlightHandlerImpl struct {
	Service FlightService
}

func (handler *FlightHandlerImpl) GetFlights(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	queryParameters := model.QueryParameters{}

	departure := query.Get("departure")
	destination := query.Get("destination")
	departureDateStr := query.Get("departure_date")
	returnDateStr := query.Get("return_date")
	airlines := query["airlines[]"]
	pageStr := query.Get("page")
	perPageStr := query.Get("per_page")
	sortDirection := query.Get("sort_direction")

	queryParameters.Carrier = airlines

	if departure == "" || destination == "" || departureDateStr == "" {
		http.Error(w, "departure, destination, departure_date and return_date are required", http.StatusBadRequest)
		return
	}
	queryParameters.Departure = departure
	queryParameters.Destination = destination

	departureDate, err := time.Parse("2006-01-02", departureDateStr)
	if err != nil {
		http.Error(w, "invalid departure_date", http.StatusBadRequest)
		return
	}
	queryParameters.DepartureDate = departureDate

	if returnDateStr != "" {
		returnDate, err := time.Parse("2006-01-02", returnDateStr)
		if err != nil {
			http.Error(w, "invalid return_date", http.StatusBadRequest)
			return
		}
		queryParameters.ReturnDate = &returnDate
	}

	if pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			http.Error(w, "invalid page", http.StatusBadRequest)
			return
		}

		if page <= 0 {
			http.Error(w, "invalid page", http.StatusBadRequest)
			return
		}
		queryParameters.CurrentPage = &page
	}

	if perPageStr != "" {
		perPage, err := strconv.Atoi(perPageStr)
		if err != nil {
			http.Error(w, "invalid per_page", http.StatusBadRequest)
			return
		}

		if perPage < 0 {
			http.Error(w, "invalid per_page", http.StatusBadRequest)
			return
		}

		queryParameters.PerPage = &perPage
	}

	if sortDirection != "" && sortDirection != "ASC" && sortDirection != "DESC" {
		http.Error(w, "invalid sort_direction", http.StatusBadRequest)
		return
	}
	queryParameters.SortDirection = sortDirection

	response, err := handler.Service.GetFlights(queryParameters)

	if err != nil {
		if errors.Is(err, service.ErrPageOutOfBounds) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		
		log.Printf("flights handler: service error: %v", err)
		http.Error(w, "unhandled error", http.StatusInternalServerError)
		return
	}
	responseDTO := toQueryResponseDTO(response)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(responseDTO); err != nil {
		log.Printf("flights handler: failed to encode response: %v", err)
	}
}

func toQueryResponseDTO(response model.QueryResponse) model.QueryResponseDTO {
	flights := make([]model.FlightDTO, 0, len(response.Flights))
	for _, flight := range response.Flights {

		var returnDate string
		if flight.ReturnDate != nil {
			returnDate = flight.ReturnDate.Format("2006-01-02")
		}
		flights = append(flights, model.FlightDTO{
			Departure:     flight.Departure,
			Destination:   flight.Destination,
			DepartureDate: flight.DepartureDate.Format("2006-01-02"),
			ReturnDate:    returnDate,
			Carrier:       flight.Carrier,
			Price:         flight.Price,
		})
	}

	airlineOffers := make([]model.AirlineOfferDTO, 0, len(response.AirlineOffers))
	for _, offer := range response.AirlineOffers {
		airlineOffers = append(airlineOffers, model.AirlineOfferDTO{
			Code:     offer.Code,
			PriceMin: offer.PriceMin,
			PriceMax: offer.PriceMax,
		})
	}

	return model.QueryResponseDTO{
		Runtime:       response.Runtime,
		FlightOffers:  response.FlightOffers,
		PageCount:     response.PageCount,
		CurrentPage:   response.CurrentPage,
		PerPage:       response.PerPage,
		AirlineOffers: airlineOffers,
		Flights:       flights,
	}
}
