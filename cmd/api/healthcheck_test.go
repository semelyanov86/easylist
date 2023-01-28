package main

import (
	"github.com/google/jsonapi"
	"net/http"
	"testing"
)

func TestHealthCheckHandler(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()
	resp, err := ts.Client().Get(ts.URL + "/api/v1/healthcheck")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("want %d status code; got %d", http.StatusOK, resp.StatusCode)
	}

	check := new(healthCheck)

	err = jsonapi.UnmarshalPayload(resp.Body, check)

	if err != nil {
		t.Fatal(err)
	}
	if check.ID != 1 {
		t.Errorf("want ID to equal 1, got %d", check.ID)
	}
	if check.Status != "available" {
		t.Errorf("want status to be available, got %s", check.Status)
	}
}
