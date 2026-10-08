package model

import (
	"time"
)

type QueryParameters struct {
	Departure     string
	Destination   string
	DepartureDate time.Time
	ReturnDate    *time.Time
	Carrier       []string
	CurrentPage   *int
	PerPage       *int
	SortDirection string
}

type FilterParameters struct {
	Departure     string
	Destination   string
	DepartureDate time.Time
	ReturnDate    *time.Time
	Carrier       []string
}

type Flight struct {
	Departure     string
	Destination   string
	DepartureDate time.Time
	ReturnDate    *time.Time
	Carrier       string
	Price         float64
}

type QueryResponse struct {
	Runtime       float64
	FlightOffers  int
	PageCount     int
	CurrentPage   int
	PerPage       int
	AirlineOffers []AirlineOffer
	Flights       []Flight
}

type QueryResponseDTO struct {
	Runtime       float64           `json:"runtime"`
	FlightOffers  int               `json:"flight_offers"`
	PageCount     int               `json:"page_count"`
	CurrentPage   int               `json:"current_page"`
	PerPage       int               `json:"per_page"`
	AirlineOffers []AirlineOfferDTO `json:"airline_offers"`
	Flights       []FlightDTO       `json:"flights"`
}

type AirlineOffer struct {
	Code     string
	PriceMin float64
	PriceMax float64
}

type AirlineOfferDTO struct {
	Code     string  `json:"code"`
	PriceMin float64 `json:"price_min"`
	PriceMax float64 `json:"price_max"`
}

type FlightDTO struct {
	Departure     string     `json:"departure"`
	Destination   string     `json:"destination"`
	DepartureDate string  `json:"date_departure"`
	ReturnDate    string `json:"date_return"`
	Carrier       string     `json:"carrier"`
	Price         float64    `json:"price"`
}
