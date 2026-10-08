package main

import (
	"fmt"
	"log"
	"net/http"

	"nitr80.com/check24-go-daemon/internal/cache"
	"nitr80.com/check24-go-daemon/internal/handler"
	"nitr80.com/check24-go-daemon/internal/loader"
	"nitr80.com/check24-go-daemon/internal/service"
)

func main() {
	paths := []string{
		"./data_csv/ama_echo_aa.csv",
		"./data_csv/ama_echo_ab.csv",
		"./data_csv/ama_echo_ac.csv",
		"./data_csv/ama_echo_ad.csv",
		"./data_csv/ama_echo_ae.csv",
		"./data_csv/ama_echo_ah.csv",
		"./data_csv/ama_echo_ag.csv",
		"./data_csv/ama_echo_ah.csv",
	}

	flights, err := loader.LoadFlights(paths)
	if err != nil {
		log.Fatalf("failed to load flights: %v", err)
	}

	cache := &cache.FlightCacheImpl{Flights: flights}
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

