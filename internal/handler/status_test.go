package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockLoadStatusProvider struct {
	loading bool
	err     error
}

func (m *mockLoadStatusProvider) GetLoadStatus() (bool, error) {
	return m.loading, m.err
}

func TestGetStatus(t *testing.T) {
	tests := []struct {
		name            string
		loading         bool
		err             error
		expectedStatus  int
		expectedLoading bool
		expectedError   bool
	}{
		{"load completed", false, nil, http.StatusOK, false, false},
		{"still loading", true, nil, http.StatusOK, true, false},
		{"load failed", false, errors.New("boom"), http.StatusServiceUnavailable, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := StatusHandlerImpl{
				Cache: &mockLoadStatusProvider{loading: tt.loading, err: tt.err},
			}
			recorder := httptest.NewRecorder()

			handler.GetStatus(recorder, httptest.NewRequest(http.MethodGet, "/status", nil))

			if recorder.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, recorder.Code)
			}

			if recorder.Header().Get("Content-Type") != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", recorder.Header().Get("Content-Type"))
			}

			var body statusResponseDTO
			if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if body.Loading != tt.expectedLoading {
				t.Errorf("expected loading %v, got %v", tt.expectedLoading, body.Loading)
			}
			if body.LoadError != tt.expectedError {
				t.Errorf("expected load_error %v, got %v", tt.expectedError, body.LoadError)
			}
		})
	}
}
