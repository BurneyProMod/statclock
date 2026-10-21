package display

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetElo(t *testing.T) {
	device := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/number/elo/set" {
			t.Errorf("path = %q, want /number/elo/set", r.URL.Path)
		}
		if got := r.URL.Query().Get("value"); got != "2009" {
			t.Errorf("value = %q, want 2009", got)
		}
	}))
	defer device.Close()

	// A trailing slash on the device URL should not matter.
	if err := SetElo(context.Background(), device.URL+"/", 2009); err != nil {
		t.Fatal(err)
	}
}

func TestSetEloError(t *testing.T) {
	device := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer device.Close()

	if err := SetElo(context.Background(), device.URL, 2009); err == nil {
		t.Error("want an error for a 404 response")
	}
}
