package handler

import (
	"encoding/json"
	"log"
	"net/http"
)

type LoadStatusProvider interface {
	GetLoadStatus() (loading bool, err error)
}

type StatusHandlerImpl struct {
	Cache LoadStatusProvider
}

type statusResponseDTO struct {
	Loading   bool `json:"loading"`
	LoadError bool `json:"load_error"`
}

func (handler *StatusHandlerImpl) GetStatus(w http.ResponseWriter, r *http.Request) {
	loading, err := handler.Cache.GetLoadStatus()

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if encodeErr := json.NewEncoder(w).Encode(statusResponseDTO{
		Loading:   loading,
		LoadError: err != nil,
	}); encodeErr != nil {
		log.Printf("status handler: failed to encode response: %v", encodeErr)
	}
}
