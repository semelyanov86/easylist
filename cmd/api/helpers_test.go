package main

import (
	"context"
	"easylist/internal/validator"
	"github.com/julienschmidt/httprouter"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"testing"
)

func TestReadIDParam(t *testing.T) {
	app := &application{}
	req := httptest.NewRequest("GET", "/path?id=123", nil)
	params := httprouter.Params{
		httprouter.Param{Key: "id", Value: "123"},
	}
	req = req.WithContext(context.WithValue(req.Context(), httprouter.ParamsKey, params))

	id, err := app.readIDParam(req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if id != 123 {
		t.Errorf("unexpected id: got %d, want 123", id)
	}

	req = httptest.NewRequest("GET", "/path?id=0", nil)
	params = httprouter.Params{
		httprouter.Param{Key: "id", Value: "0"},
	}
	req = req.WithContext(context.WithValue(req.Context(), httprouter.ParamsKey, params))

	id, err = app.readIDParam(req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if id != 0 {
		t.Errorf("unexpected id: got %d, want 0", id)
	}

	req = httptest.NewRequest("GET", "/path?id=invalid", nil)
	params = httprouter.Params{
		httprouter.Param{Key: "id", Value: "invalid"},
	}
	req = req.WithContext(context.WithValue(req.Context(), httprouter.ParamsKey, params))

	id, err = app.readIDParam(req)
	if id != 0 {
		t.Errorf("expected id is 0, but %d", id)
	}
}

func TestSaveFile(t *testing.T) {
	app := &application{}

	file := "VGhpcyBpcyBhIHRlc3QgdGVzdCBmaWxl"
	userId := int64(123)

	// Test with valid file
	actualFileName, err := app.saveFile(file, userId)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = os.Stat(actualFileName)
	if os.IsNotExist(err) {
		t.Errorf("file not created: %s", actualFileName)
	}
	err = os.Remove(actualFileName)
	if err != nil {
		t.Fatal(err)
	}

	// Test with empty file
	actualFileName, err = app.saveFile("", userId)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if actualFileName != "" {
		t.Errorf("unexpected file name: got %s, want \"\"", actualFileName)
	}

	// Test with invalid base64 string
	actualFileName, err = app.saveFile("invalid", userId)
	if err == nil {
		t.Error("expected error, but got nil")
	}
	if actualFileName != "" {
		t.Errorf("unexpected file name: got %s, want \"\"", actualFileName)
	}
}

func TestReadString(t *testing.T) {
	app := &application{}

	qs := url.Values{}
	qs.Set("name", "Alice")

	// Test with existing key
	expectedValue := "Alice"
	actualValue := app.readString(qs, "name", "Bob")
	if actualValue != expectedValue {
		t.Errorf("unexpected value: got %s, want %s", actualValue, expectedValue)
	}

	// Test with missing key
	expectedValue = "Bob"
	actualValue = app.readString(qs, "email", "Bob")
	if actualValue != expectedValue {
		t.Errorf("unexpected value: got %s, want %s", actualValue, expectedValue)
	}

	// Test with empty default value
	expectedValue = "Alice"
	actualValue = app.readString(qs, "name", "")
	if actualValue != expectedValue {
		t.Errorf("unexpected value: got %s, want %s", actualValue, expectedValue)
	}

	// Test with empty string value
	qs.Set("name", "")
	expectedValue = "Bob"
	actualValue = app.readString(qs, "name", "Bob")
	if actualValue != expectedValue {
		t.Errorf("unexpected value: got %s, want %s", actualValue, expectedValue)
	}
}

func TestReadCSV(t *testing.T) {
	app := &application{}

	qs := url.Values{}
	qs.Set("tags", "food,travel")

	// Test with existing key
	expectedValue := []string{"food", "travel"}
	actualValue := app.readCSV(qs, "tags", []string{"music"})
	if !reflect.DeepEqual(actualValue, expectedValue) {
		t.Errorf("unexpected value: got %v, want %v", actualValue, expectedValue)
	}

	// Test with missing key
	expectedValue = []string{"music"}
	actualValue = app.readCSV(qs, "colors", []string{"music"})
	if !reflect.DeepEqual(actualValue, expectedValue) {
		t.Errorf("unexpected value: got %v, want %v", actualValue, expectedValue)
	}

	// Test with empty default value
	expectedValue = []string{"food", "travel"}
	actualValue = app.readCSV(qs, "tags", []string{})
	if !reflect.DeepEqual(actualValue, expectedValue) {
		t.Errorf("unexpected value: got %v, want %v", actualValue, expectedValue)
	}

	// Test with empty CSV string value
	qs.Set("tags", "")
	expectedValue = []string{"music"}
	actualValue = app.readCSV(qs, "tags", []string{"music"})
	if !reflect.DeepEqual(actualValue, expectedValue) {
		t.Errorf("unexpected value: got %v, want %v", actualValue, expectedValue)
	}
}

func TestReadInt(t *testing.T) {
	app := &application{}
	v := validator.New()

	qs := url.Values{}
	qs.Set("page", "3")

	// Test with existing key and valid integer value
	expectedValue := 3
	actualValue := app.readInt(qs, "page", 1, v)
	if actualValue != expectedValue {
		t.Errorf("unexpected value: got %d, want %d", actualValue, expectedValue)
	}

	// Test with existing key and invalid integer value
	expectedValue = 1
	qs.Set("page", "inv")
	actualValue = app.readInt(qs, "page", 1, v)
	if actualValue != expectedValue {
		t.Errorf("unexpected value: got %d, want %d", actualValue, expectedValue)
	}

	// Test with missing key
	expectedValue = 2
	actualValue = app.readInt(qs, "limit", 2, v)
	if actualValue != expectedValue {
		t.Errorf("unexpected value: got %d, want %d", actualValue, expectedValue)
	}

	// Test with empty default value
	expectedValue = 0
	qs.Set("page", "")
	actualValue = app.readInt(qs, "page", 0, v)
	if actualValue != expectedValue {
		t.Errorf("unexpected value: got %d, want %d", actualValue, expectedValue)
	}
}

func TestReadBool(t *testing.T) {
	app := &application{}
	v := validator.New()

	qs := url.Values{}
	qs.Set("is_active", "true")

	// Test with existing key and valid bool value
	expectedValue := true
	actualValue := app.readBool(qs, "is_active", false, v)
	if actualValue != true {
		t.Errorf("unexpected value: got %t, want %t", actualValue, true)
	}

	// Test with existing key and invalid bool value
	expectedValue = false
	qs.Set("is_active", "inv")
	actualValue = app.readBool(qs, "is_active", false, v)
	if actualValue != expectedValue {
		t.Errorf("unexpected value: got %t, want %t", actualValue, expectedValue)
	}

	// Test with missing key
	expectedValue = true
	actualValue = app.readBool(qs, "is_deleted", true, v)
	if actualValue != expectedValue {
		t.Errorf("unexpected value: got %t, want %t", actualValue, expectedValue)
	}

	// Test with empty default value
	expectedValue = false
	actualValue = app.readBool(qs, "is_active", false, v)
	if actualValue != expectedValue {
		t.Errorf("unexpected value: got %t, want %t", actualValue, expectedValue)
	}
}
