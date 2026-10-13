package cache

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"nitr80.com/check24-go-daemon/internal/model"
)

type FlightCacheImpl struct {
	Mu      sync.RWMutex
	Flights map[string][]model.Flight
}

func NewFlightCacheImpl() *FlightCacheImpl {
	return &FlightCacheImpl{Flights: make(map[string][]model.Flight)}
}

func (cache *FlightCacheImpl) GetFlights(filterParameters model.FilterParameters) []model.Flight {
	key := createKeyWithFilterParameters(filterParameters)
	// fmt.Println(key)
	return cache.Flights[key]
}

var ErrFlightNotInDE = errors.New("flight is not in german market")
var ErrNoAdult = errors.New("flight does not have any adult")

func (cache *FlightCacheImpl) LoadFlights(paths []string) error {
	var wg sync.WaitGroup

	results := make(chan model.Flight, len(paths)*100)
	errs := make(chan error, len(paths))

	for _, path := range paths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			err := readFlights(path, results)
			if err != nil {
				errs <- fmt.Errorf("reading %s: %w", path, err)
				return
			}
		}(path)
	}

	go func() {
		wg.Wait()
		close(results)
		close(errs)
	}()

	for flight := range results {
		key := createKeyWithFlight(flight)
		flights, ok := cache.Flights[key]
		if !ok {
			cache.Mu.Lock()
			newFlights := make([]model.Flight, 0)
			newFlights = append(newFlights, flight)
			cache.Flights[key] = newFlights
			cache.Mu.Unlock()
		} else {
			cache.Mu.Lock()
			flights = append(flights, flight)
			cache.Flights[key] = flights
			cache.Mu.Unlock()
		}
	}

	var firstErr error
	for err := range errs {
		if firstErr == nil {
			firstErr = err
		}
	}

	if firstErr != nil {
		return firstErr
	}

	return nil
}

func createKeyWithFlight(flight model.Flight) string {
	departureDateString := flight.DepartureDate.Format(time.DateOnly)
	var returnDateString string
	if flight.ReturnDate != nil {
		returnDateString = flight.ReturnDate.Format(time.DateOnly)
	}
	key := fmt.Sprintf("%v|%v|%v|%v", flight.Departure, flight.Destination, departureDateString, returnDateString)

	return key
}

func createKeyWithFilterParameters(filterParameters model.FilterParameters) string {
	departureDateString := filterParameters.DepartureDate.Format(time.DateOnly)
	var returnDateString string
	if filterParameters.ReturnDate != nil {
		returnDateString = filterParameters.ReturnDate.Format(time.DateOnly)
	}
	key := fmt.Sprintf("%v|%v|%v|%v", filterParameters.Departure, filterParameters.Destination, departureDateString, returnDateString)

	return key
}

func readFlights(filename string, results chan<- model.Flight) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = '^'
	reader.FieldsPerRecord = -1

	for {
		record, err := reader.Read()
		// fmt.Printf("Record len: %v", len(record))

		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		flight, err := parseFlight(record)
		if err != nil {
			if errors.Is(err, ErrFlightNotInDE) || errors.Is(err, ErrNoAdult) {
				continue
			} else {
				return err
			}
		}

		results <- flight
	}

	return nil
}

func parseFlight(record []string) (model.Flight, error) {
	if len(record) != 133 {
		return model.Flight{}, fmt.Errorf("invalid record: expected at least 133 fields, got %d", len(record))
	}

	if record[2] != "DE" {
		return model.Flight{}, ErrFlightNotInDE
	}

	departureDate, err := time.Parse(
		"2006-01-02",
		record[7],
	)
	if err != nil {
		return model.Flight{}, fmt.Errorf("invalid departure date: %w", err)
	}

	var returnDate *time.Time

	if record[8] != "" {
		parsed, err := time.Parse("2006-01-02", record[8])
		if err != nil {
			return model.Flight{}, fmt.Errorf("invalid return date: %w", err)
		}

		returnDate = &parsed
	}

	price, err := strconv.ParseFloat(record[12], 64)
	if err != nil {
		return model.Flight{}, fmt.Errorf("invalid price: %w", err)
	}

	passengers, err := parsePassengerInfo(record[10])
	if err != nil {
		return model.Flight{}, err
	}

	adults := passengers["ADT"]
	// children := passengers["CHI"]

	adults, ok := passengers["ADT"]
	if !ok || adults <= 0 {
		return model.Flight{}, ErrNoAdult
	}

	price /= float64(adults)

	return model.Flight{
		Departure:     record[5],
		Destination:   record[6],
		DepartureDate: departureDate,
		ReturnDate:    returnDate,
		Carrier:       record[132],
		Price:         price,
	}, nil
}

func parsePassengerInfo(value string) (map[string]int, error) {
	result := make(map[string]int)

	if value == "" {
		return result, nil
	}

	for _, part := range strings.Split(value, ",") {
		keyValue := strings.SplitN(part, "=", 2)

		if len(keyValue) != 2 {
			return nil, fmt.Errorf("invalid passenger info: %q", part)
		}

		count, err := strconv.Atoi(keyValue[1])
		if err != nil {
			return nil, fmt.Errorf("invalid passenger count for %q: %w", keyValue[0], err)
		}

		result[keyValue[0]] = count
	}

	return result, nil
}
