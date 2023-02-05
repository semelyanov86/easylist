package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"testing"
)

const WrongToken = "wrong_token"

func TestAuthorization(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	defer ts.Close()
	user, _, err := createTestUserWithToken(t, app, "")
	folder, err := createTestFolder(app, user.ID, "", 0)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		urlPath string
		method  string
	}{
		{"GetFolderById", ts.URL + "/api/v1/folders/" + strconv.Itoa(int(folder.ID)), "GET"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := generateRequestWithToken(tt.urlPath, WrongToken, tt.method, nil)
			resp, err := ts.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("want %d status code; got %d", http.StatusUnauthorized, resp.StatusCode)
			}
		})
	}
}

func TestNotOwnedRecordAccess(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	defer ts.Close()
	user, _, err := createTestUserWithToken(t, app, "")
	folder, err := createTestFolder(app, user.ID, "", 0)

	_, token, err := createTestUserWithToken(t, app, "test2@mail.ru")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		urlPath string
		method  string
	}{
		{"GetFolderById", ts.URL + "/api/v1/folders/" + strconv.Itoa(int(folder.ID)), "GET"},
		{"DeleteFolderById", ts.URL + "/api/v1/folders/" + strconv.Itoa(int(folder.ID)), "DELETE"},
	}
	var errorData JsonapiErrors

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := generateRequestWithToken(tt.urlPath, token.Plaintext, tt.method, nil)
			resp, err := ts.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			if resp.StatusCode != http.StatusNotFound {
				t.Errorf("want %d status code; got %d", http.StatusNotFound, resp.StatusCode)
			}
			err = json.Unmarshal(body, &errorData)
			if err != nil {
				t.Fatal(err)
			}
			if errorData.Errors[0].Title != "Not Found" {
				t.Errorf("want error title to be %s; got %s", "Not Found", errorData.Errors[0].Title)
			}
		})
	}
}
