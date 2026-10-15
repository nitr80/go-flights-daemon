package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"nitr80.com/flights-daemon/internal/cache"
	"nitr80.com/flights-daemon/internal/handler"
	"nitr80.com/flights-daemon/internal/service"
)

func main() {

	dir := os.Getenv("DIR_PATH")

	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalf("failed read dir: %v", err)
	}
	var paths []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".csv" {
			paths = append(paths, filepath.Join(dir, entry.Name()))
		}
	}

	cache := cache.NewFlightCacheImpl()

	go func() {
		err = cache.LoadFlights(paths)

		if err != nil {
			log.Printf("WARNING: flight data loading was incomplete: %v", err)
		}
	}()

	service := &service.FlightServiceImpl{Cache: cache}
	handler := &handler.FlightHandlerImpl{Service: service}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /flights", handler.GetFlights)

	fmt.Println("Serving and listening on :8080")

	err = http.ListenAndServe(":8080", mux)

	if err != nil {
		log.Fatalf("failed to listen and serve: %v", err)
	}
}

// Notes
// 1 more hour i could implement optionals
// data paths strings slice could be more clever
// pagination and sorting in service
// filtering in cache
// I could have passed the request context and add a timeout
// mapping is not so useful but i still makes sense
// could have logged more and error handling could have been better (with wrapping)
// I forgot to add constructor and it was a little late to add, but also not a big problem
// caching between stops
