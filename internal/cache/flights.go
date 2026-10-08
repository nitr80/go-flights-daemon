package cache

import (

	"nitr80.com/check24-go-daemon/internal/model"
)

type FlightCacheImpl struct {
	Flights []model.Flight
}

func (cache *FlightCacheImpl) GetFlights(filterParameters model.FilterParameters) []model.Flight {
	filterFlights := make([]model.Flight, 0)

	for _, flight := range cache.Flights {

		if filterParameters.Departure == flight.Departure &&
			filterParameters.Destination == flight.Destination &&
			filterParameters.DepartureDate.Equal(flight.DepartureDate) {
			if filterParameters.ReturnDate != nil {
				if flight.ReturnDate != nil &&
					filterParameters.ReturnDate.Equal(*flight.ReturnDate) {
					filterFlights = append(filterFlights, flight)
				}
			} else if flight.ReturnDate == nil {
				filterFlights = append(filterFlights, flight)
			}
		}
	}

	return filterFlights
}
