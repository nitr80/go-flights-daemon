package service

import "nitr80.com/check24-go-daemon/internal/model"

type FlightCache interface {
	GetFlights(filterParameters model.FilterParameters) []model.Flight
}

type FlightServiceImpl struct {
	cache FlightCache
}

// continue service, cache and handler; decide where to map the dto
func (service *FlightServiceImpl) GetFlights(queryParameters model.QueryParameters) (model.QueryResponse, error) {
	filterParameters := model.FilterParameters{
		Departure:     queryParameters.Departure,
		Destination:   queryParameters.Destination,
		DepartureDate: queryParameters.DepartureDate,
		ReturnDate:    queryParameters.ReturnDate,
		Carrier:       queryParameters.Carrier,
	}

	flights := service.cache.GetFlights(filterParameters)

	return model.QueryResponse{Flights: flights}, nil
}
