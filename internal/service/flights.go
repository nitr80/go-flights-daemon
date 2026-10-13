package service

import (
	"time"

	"nitr80.com/check24-go-daemon/internal/model"
)

type FlightCache interface {
	GetFlights(filterParameters model.FilterParameters) []model.Flight
	// LoadFlights(paths []string) error
}

type FlightServiceImpl struct {
	Cache FlightCache
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

	start := time.Now()
	flights := service.Cache.GetFlights(filterParameters)
	runtime := time.Since(start)

	return model.QueryResponse{
		Flights:       flights,
		Runtime:       runtime.Seconds(),
		FlightOffers:  len(flights),
		PageCount:     0,
		CurrentPage:   0,
		PerPage:       0,
		AirlineOffers: createAirlineOffers(flights),
	}, nil
}

func createAirlineOffers(flights []model.Flight) []model.AirlineOffer {
	offers := make(map[string]*model.AirlineOffer)

	for _, flight := range flights {
		offer, exists := offers[flight.Carrier]

		if !exists {
			offers[flight.Carrier] = &model.AirlineOffer{
				Code:     flight.Carrier,
				PriceMin: flight.Price,
				PriceMax: flight.Price,
			}
			continue
		}

		if offer.PriceMax < flight.Price {
			offer.PriceMax = flight.Price
		}

		if offer.PriceMin > flight.Price {
			offer.PriceMin = flight.Price
		}
	}

	result := make([]model.AirlineOffer, 0, len(offers))

	for _, offer := range offers {
		result = append(result, *offer)
	}

	return result
}
