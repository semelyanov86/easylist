package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestTokenSuccessfullyCreated(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	_, token, err := createTestUserWithToken(t, app, "test@mail.ru")
	if err != nil {
		t.Fatal(err)
	}
	defer ts.Close()

	var tokenData = []byte(`{
	  "data": {
		"type": "tokens",
		"attributes": {
		  "email": "test@mail.ru",
				"password": "password123"
		}
	  }
	}`)

	req := generateRequestWithToken(ts.URL+"/api/v1/tokens/authentication", token.Plaintext, "POST", bytes.NewBuffer(tokenData))
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("want %d status code; got %d", http.StatusCreated, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	var check = Input[ActivationAttributes]{
		Data: InputAttributes[ActivationAttributes]{},
	}
	err = json.Unmarshal(body, &check)

	if err != nil {
		t.Fatal(err)
	}

	if resp.Header.Get("Content-Type") != "application/vnd.api+json" {
		t.Errorf("want Content-Type to be application/vnd.api+json, got %s", resp.Header.Get("Content-Type"))
	}
	if check.Data.Type != "tokens" {
		t.Errorf("expected error type tokens, got %s", check.Data.Type)
	}
	if check.Data.Attributes.Token == "" {
		t.Error("Token is empty")
	}
}

func TestTokenValidation(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	defer ts.Close()
	_, token, err := createTestUserWithToken(t, app, "test2@mail.ru")

	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		content string
	}{
		{
			name: "data.attributes.email",
			content: `{
			  "data": {
				"type": "tokens",
				"attributes": {
				  "email": "",
						"password": "passwordpassword"
				}
			  }
			}`,
		},
		{
			name: "data.attributes.password",
			content: `{
			  "data": {
				"type": "tokens",
				"attributes": {
				  "email": "test2@mail.ru",
						"password": "123"
				}
			  }
			}`,
		},
	}

	var errorData JsonapiErrors

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := generateRequestWithToken(ts.URL+"/api/v1/tokens/authentication", token.Plaintext, "POST", bytes.NewBuffer([]byte(tt.content)))
			resp, err := ts.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusUnprocessableEntity {
				t.Errorf("want %d status code; got %d", http.StatusUnprocessableEntity, resp.StatusCode)
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			err = json.Unmarshal(body, &errorData)
			if err != nil {
				t.Fatal(err)
			}
			if errorData.Errors[0].Title != "Validation failed for field "+tt.name {
				t.Errorf("want error title to be %s; got %s", "Validation failed for field", errorData.Errors[0].Title)
			}

		})
	}
}

func TestTokenNotSetIfCredentialsAreWrong(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	defer ts.Close()
	_, token, err := createTestUserWithToken(t, app, "test2@mail.ru")

	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		content string
	}{
		{
			name: "data.attributes.email",
			content: `{
			  "data": {
				"type": "tokens",
				"attributes": {
				  "email": "test2@mail.ru",
						"password": "password12rfdsfsd"
				}
			  }
			}`,
		},
		{
			name: "data.attributes.password",
			content: `{
			  "data": {
				"type": "tokens",
				"attributes": {
				  "email": "test2@mai33l.ru",
						"password": "password123"
				}
			  }
			}`,
		},
	}

	var errorData JsonapiErrors

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := generateRequestWithToken(ts.URL+"/api/v1/tokens/authentication", token.Plaintext, "POST", bytes.NewBuffer([]byte(tt.content)))
			resp, err := ts.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("want %d status code; got %d", http.StatusUnauthorized, resp.StatusCode)
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			err = json.Unmarshal(body, &errorData)
			if err != nil {
				t.Fatal(err)
			}
			if errorData.Errors[0].Title != "Auth Error" {
				t.Errorf("want error title to be %s; got %s", "Auth Error", errorData.Errors[0].Title)
			}

		})
	}
}
