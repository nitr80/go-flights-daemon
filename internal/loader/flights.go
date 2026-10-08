package loader

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

var ErrFlightNotInDE = errors.New("flight is not in german market")

func LoadFlights(paths []string) ([]model.Flight, error) {
	var wg sync.WaitGroup

	results := make(chan []model.Flight, len(paths))
	errs := make(chan error, len(paths))

	for _, path := range paths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			flights, err := readFlights(path)
			if err != nil {
				errs <- fmt.Errorf("reading %s: %w", path, err)
				return
			}
			results <- flights
		}(path)
	}

	wg.Wait()

	close(results)
	close(errs)
	for err := range errs {
		if err != nil {
			return nil, err
		}
	}
	var allFlights []model.Flight
	for flights := range results {
		allFlights = append(allFlights, flights...)
	}
	return allFlights, nil
}

func readFlights(filename string) ([]model.Flight, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = '^'
	reader.FieldsPerRecord = -1

	var flights []model.Flight

	for {
		record, err := reader.Read()
		// fmt.Printf("Record len: %v", len(record))

		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}

		flight, err := parseFlight(record)
		if err != nil {
			if errors.Is(err, ErrFlightNotInDE) {
				continue
			} else {
				return nil, err
			}
		}

		flights = append(flights, flight)
	}

	return flights, nil
}

func parseFlight(record []string) (model.Flight, error) {
	if len(record) != 133 {
		return model.Flight{}, fmt.Errorf("invalid record: expected at least 13 fields, got %d", len(record))
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
