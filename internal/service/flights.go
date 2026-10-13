package service

import (
	"cmp"
	"errors"
	"slices"
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

var ErrPageOutOfBounds = errors.New("current page is greater than page count")

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

	switch queryParameters.SortDirection {
	case "", "ASC":
		slices.SortFunc(flights, func(a, b model.Flight) int {
			return cmp.Compare(a.Price, b.Price)
		})
	case "DESC":
		slices.SortFunc(flights, func(a, b model.Flight) int {
			return cmp.Compare(b.Price, a.Price)
		})
	}

	if queryParameters.CurrentPage != nil && queryParameters.PerPage != nil {
		pageCount := len(flights) / *queryParameters.PerPage
		if len(flights) % *queryParameters.PerPage != 0 {
			pageCount++
		}

		if pageCount < *queryParameters.CurrentPage {
			return model.QueryResponse{}, ErrPageOutOfBounds
		}

		start := (*queryParameters.CurrentPage - 1) * *queryParameters.PerPage
		end := start + *queryParameters.PerPage
		if end > len(flights) {
			end = len(flights)
		}

		pagedFlights := flights[start : end]

		return model.QueryResponse{
			Flights:       pagedFlights,
			Runtime:       runtime.Seconds(),
			FlightOffers:  len(flights),
			PageCount:     pageCount,
			CurrentPage:   *queryParameters.CurrentPage,
			PerPage:       *queryParameters.PerPage,
			AirlineOffers: createAirlineOffers(flights),
		}, nil

	}

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
