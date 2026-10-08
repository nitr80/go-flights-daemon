package loader

import (
	"testing"
)

func TestReadFlights(t *testing.T) {
	flights, err := readFlights("testdata/test1.csv")
	if err != nil {
		t.Fatal(err)
	}

	if len(flights) != 4 {
		t.Fatalf("expected 4 flights, got %d", len(flights))
	}

	got := flights[0]

	if got.Departure != "CGN" {
		t.Errorf("expected departure CGN, got %s", got.Departure)
	}

	if got.Destination != "STR" {
		t.Errorf("expected destination STR, got %s", got.Destination)
	}

	if got.Price != 234.25 {
		t.Errorf("expected price 234.25, got %f", got.Price)
	}

	if got.Carrier != "LH" {
		t.Errorf("expected carrier LH, got %s", got.Carrier)
	}
}

func TestLoadFlights(t *testing.T) {
	paths := []string{
		"testdata/test1.csv",
		"testdata/test2.csv",
	}

	flights, err := LoadFlights(paths)
	if err != nil {
		t.Fatalf("LoadFlights() error = %v", err)
	}

	if len(flights) != 15 {
		t.Fatalf("expected 15 flights, got %d", len(flights))
	}
}
