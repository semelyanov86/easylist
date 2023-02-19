package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestActivationHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/activate", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(activation)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expectedContentType := "text/html; charset=utf-8"
	if ct := rr.Header().Get("Content-Type"); ct != expectedContentType {
		t.Errorf("handler returned wrong content type: got %v want %v", ct, expectedContentType)
	}

	expectedBody := "Activate account"

	if body := rr.Body.String(); !strings.Contains(body, expectedBody) {
		t.Errorf("handler returned unexpected body: got %v want %v", body, expectedBody)
	}
}
