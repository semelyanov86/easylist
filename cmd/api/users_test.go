package main

import (
	"bytes"
	"easylist/internal/data"
	"encoding/json"
	"errors"
	"github.com/google/jsonapi"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}(resp.Body)

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
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					t.Fatal(err)
				}
			}(resp.Body)

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

func TestUserUpdate(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	user, token, err := createTestUserWithToken(t, app, "")
	if err != nil {
		t.Fatal(err)
	}
	defer ts.Close()
	var userData = []byte(`{
	  "data": {
		"type": "users",
			"id": "` + strconv.Itoa(int(user.ID)) + `",
		"attributes": {
		  "name": "Sergey Emelyanov",
		  "email": "emelyanov86@km.ru",
				"password": "SuperPasswordHere"
		}
	  }
	}`)

	req := generateRequestWithToken(ts.URL+"/api/v1/users/"+strconv.Itoa(int(user.ID)), token.Plaintext, "PATCH", bytes.NewBuffer(userData))
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("want %d status code; got %d", http.StatusOK, resp.StatusCode)
	}

	check := new(data.User)

	err = jsonapi.UnmarshalPayload(resp.Body, check)

	if err != nil {
		t.Fatal(err)
	}
	if check.ID != user.ID {
		t.Errorf("want ID to be %d, got %d", user.ID, check.ID)
	}
	if check.Name != "Sergey Emelyanov" {
		t.Errorf("want Name to be %s, got %s", "Sergey Emelyanov", check.Name)
	}
	if check.Email != "emelyanov86@km.ru" {
		t.Errorf("want Email to be %s, got %s", "emelyanov86@km.ru", check.Email)
	}

	newUser, err := app.models.Users.GetByEmail("emelyanov86@km.ru")
	if err != nil {
		t.Fatal(err)
	}
	if !newUser.Password.Matches("SuperPasswordHere") {
		t.Errorf("New password do not match to value SuperPasswordHere")
	}
}

func TestShowCurrentUser(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	user, token, err := createTestUserWithToken(t, app, "")
	if err != nil {
		t.Fatal(err)
	}
	defer ts.Close()
	req := generateRequestWithToken(ts.URL+"/api/v1/my", token.Plaintext, "", nil)
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("want %d status code; got %d", http.StatusOK, resp.StatusCode)
	}

	check := new(data.User)

	err = jsonapi.UnmarshalPayload(resp.Body, check)

	if err != nil {
		t.Fatal(err)
	}
	if check.ID != user.ID {
		t.Errorf("want ID to be %d, got %d", user.ID, check.ID)
	}
	if check.Name != user.Name {
		t.Errorf("want name to be %s, got %s", user.Name, check.Name)
	}
	if check.Email != user.Email {
		t.Errorf("want name to be %s, got %s", user.Email, check.Email)
	}
}

func TestDeleteUserWithAllContent(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	user, token, err := createTestUserWithToken(t, app, "")
	if err != nil {
		t.Fatal(err)
	}
	defer ts.Close()

	folder, err := createTestFolder(app, user.ID, "Adv name", 0)
	if err != nil {
		t.Fatal(err)
	}

	req := generateRequestWithToken(ts.URL+"/api/v1/users/"+strconv.Itoa(int(user.ID)), token.Plaintext, "DELETE", nil)
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("want %d status code; got %d", http.StatusNoContent, resp.StatusCode)
	}
	oldUser, _ := app.models.Users.GetByEmail(user.Email)

	if oldUser != nil {
		t.Errorf("Expected no user, got user with id %d", oldUser.ID)
	}

	oldFolder, err := app.models.Folders.Get(folder.ID, user.ID)
	if oldFolder != nil {
		t.Errorf("Expected no folder, got folder with id %d", oldFolder.ID)
	}
	if !errors.Is(err, data.ErrRecordNotFound) {
		t.Errorf("Expected record not found error, got %s", err.Error())
	}
}

func TestUpdatePasswordWithTokenHandler(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	user, _, err := createTestUserWithToken(t, app, "")
	if err != nil {
		t.Fatal(err)
	}
	defer ts.Close()
	token, err := app.models.Tokens.New(user.ID, 45*time.Minute, data.ScopePasswordReset)
	if err != nil {
		t.Fatal(err)
		return
	}
	var tokenData = []byte(`{
	  "data": {
		"type": "tokens",
		"attributes": {
		  "token": "` + token.Plaintext + `",
				"password": "SuperPasswordHere"
		}
	  }
	}`)

	req := generateRequestWithToken(ts.URL+"/api/v1/users/password", "", "PUT", bytes.NewBuffer(tokenData))
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("want %d status code; got %d", http.StatusOK, resp.StatusCode)
	}

	check := new(data.User)

	err = jsonapi.UnmarshalPayload(resp.Body, check)

	if err != nil {
		t.Fatal(err)
	}
	if check.ID != user.ID {
		t.Errorf("want ID to be %d, got %d", user.ID, check.ID)
	}
}
