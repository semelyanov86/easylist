package main

import (
	"bytes"
	"easylist/internal/data"
	"encoding/json"
	"github.com/google/jsonapi"
	"io"
	"net/http"
	"testing"
)

func TestUserRegistration(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	app.config.Registration = true
	defer teardown()

	ts := newTestServer(t, app.routes())

	var userData = []byte(`{
	  "data": {
		"type": "users",
		"attributes": {
		  "name": "Sergey Emelyanov",
		  "email": "emelyanov86@km.ru",
           "password": "passpasspass"
		}
	  }
	}`)
	req := generateRequestWithToken(ts.URL+"/api/v1/users/", "", "POST", bytes.NewBuffer(userData))
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("want %d status code; got %d", http.StatusCreated, resp.StatusCode)
	}

	check := new(data.User)

	err = jsonapi.UnmarshalPayload(resp.Body, check)

	if err != nil {
		t.Fatal(err)
	}
	if check.ID < 1 {
		t.Errorf("want correct ID, got %d", check.ID)
	}
	if check.Name != "Sergey Emelyanov" {
		t.Errorf("want name to be Sergey Emelyanov, got %s", check.Name)
	}
	if check.Email != "emelyanov86@km.ru" {
		t.Errorf("want email to be emelyanov86@km.ru, got %s", check.Email)
	}
}

func TestUserValidation(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	app.config.Registration = true
	defer teardown()

	ts := newTestServer(t, app.routes())
	defer ts.Close()
	_, _, err := createTestUserWithToken(t, app, "")

	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		content string
	}{
		{
			name: "data.attributes.name",
			content: `{
			  "data": {
				"type": "users",
				"attributes": {
				  "name": "",
				  "email": "emelyanov86@km.ru",
				   "password": "passpasspass"
				}
			  }
			}`,
		},
		{
			name: "data.attributes.email",
			content: `{
			  "data": {
				"type": "users",
				"attributes": {
				  "name": "Sergey Emelyanov",
				  "email": "",
				   "password": "passpasspass"
				}
			  }
			}`,
		},
		{
			name: "data.attributes.email",
			content: `{
			  "data": {
				"type": "users",
				"attributes": {
				  "name": "Sergey Emelyanov",
				  "email": "wrong-email",
				   "password": "passpasspass"
				}
			  }
			}`,
		},
		{
			name: "data.attributes.password",
			content: `{
			  "data": {
				"type": "users",
				"attributes": {
				  "name": "Sergey Emelyanov",
				  "email": "emelyanov86@list.ru",
				   "password": "short"
				}
			  }
			}`,
		},
		{
			name: "data.attributes.email",
			content: `{
			  "data": {
				"type": "users",
				"attributes": {
				  "name": "Sergius Polov",
				  "email": "test@mail.ru",
				   "password": "CorrectPasswordLong"
				}
			  }
			}`,
		},
	}

	var errorData JsonapiErrors

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := generateRequestWithToken(ts.URL+"/api/v1/users", "", "POST", bytes.NewBuffer([]byte(tt.content)))
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
				t.Errorf("want error title to be %s; got %s", "Validation failed for field "+tt.name, errorData.Errors[0].Title)
			}

		})
	}
}
