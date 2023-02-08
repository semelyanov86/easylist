package main

import (
	"bytes"
	"easylist/internal/data"
	"encoding/json"
	"github.com/google/jsonapi"
	"io"
	"net/http"
	"strconv"
	"testing"
)

func TestShowNewCreatedList(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	user, token, err := createTestUserWithToken(t, app, "")
	if err != nil {
		t.Fatal(err)
	}
	defer ts.Close()
	folder, err := createTestFolder(app, user.ID, "", 0)
	if err != nil {
		t.Fatal(err)
	}
	var list = data.List{}
	list.FolderId = folder.ID
	list.UserId = user.ID
	err = createTestList(app, &list)
	if err != nil {
		t.Fatal(err)
	}
	req := generateRequestWithToken(ts.URL+"/api/v1/lists/"+strconv.Itoa(int(list.ID)), token.Plaintext, "", nil)
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("want %d status code; got %d", http.StatusOK, resp.StatusCode)
	}

	check := new(data.List)

	err = jsonapi.UnmarshalPayload(resp.Body, check)

	if err != nil {
		t.Fatal(err)
	}
	if check.ID != list.ID {
		t.Errorf("want ID to equal %d, got %d", list.ID, check.ID)
	}
	if check.Name != list.Name {
		t.Errorf("want Name to be %s, got %s", list.Name, check.Name)
	}
	if check.Icon != list.Icon {
		t.Errorf("want Icon to be %s, got %s", list.Icon, check.Icon)
	}
}

func TestListNotFoundAccess(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	defer ts.Close()
	user, token, err := createTestUserWithToken(t, app, "")
	err = createTestList(app, &data.List{UserId: user.ID})

	if err != nil {
		t.Fatal(err)
	}
	tests := []string{
		"adfad", "4", "-2",
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			req := generateRequestWithToken(ts.URL+"/api/v1/lists/"+tt, token.Plaintext, "", nil)
			resp, err := ts.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusNotFound {
				t.Errorf("want %d status code; got %d", http.StatusNotFound, resp.StatusCode)
			}
		})
	}
}

func TestListCreate(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	user, token, err := createTestUserWithToken(t, app, "")
	if err != nil {
		t.Fatal(err)
	}
	folder, err := createTestFolder(app, user.ID, "", 0)
	defer ts.Close()

	var listData = []byte(`{
	  "data": {
		"type": "lists",
		"attributes": {
		  "name": "Some Test List",
		  "icon": "fa-list",
           "folder_id": ` + strconv.Itoa(int(folder.ID)) + `
		}
	  }
	}`)
	req := generateRequestWithToken(ts.URL+"/api/v1/lists", token.Plaintext, "POST", bytes.NewBuffer(listData))
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("want %d status code; got %d", http.StatusCreated, resp.StatusCode)
	}

	check := new(data.List)

	err = jsonapi.UnmarshalPayload(resp.Body, check)

	if err != nil {
		t.Fatal(err)
	}
	if check.ID < 1 {
		t.Errorf("want correct ID, got %d", check.ID)
	}
	if check.Name != "Some Test List" {
		t.Errorf("want Name to be %s, got %s", "New testing folder", check.Name)
	}
	if check.Icon != "fa-list" {
		t.Errorf("want Icon to be fa-some, got %s", check.Icon)
	}
	if check.Order != 1 {
		t.Errorf("want Order to be 1, got %d", check.Order)
	}
	if check.FolderId != folder.ID {
		t.Errorf("want folderid to be %d, got %d", folder.ID, check.FolderId)
	}
	if resp.Header.Get("Location") != "http://127.0.0.1/api/v1/lists/"+strconv.Itoa(int(check.ID)) {
		t.Errorf("want location header with folders value, got %s", resp.Header.Get("Location"))
	}
	if resp.Header.Get("Content-Type") != "application/vnd.api+json" {
		t.Errorf("want Content-Type to be application/vnd.api+json, got %s", resp.Header.Get("Content-Type"))
	}
}

func TestListValidation(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	defer ts.Close()
	user, token, err := createTestUserWithToken(t, app, "")
	folder, err := createTestFolder(app, user.ID, "", 0)

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
		"type": "lists",
		"attributes": {
		  "name": "",
		  "icon": "fa-list",
           "folder_id": ` + strconv.Itoa(int(folder.ID)) + `
		}
	  }
	}`,
		},
		{
			name: "data.attributes.icon",
			content: `{
	  "data": {
		"type": "lists",
		"attributes": {
		  "name": "Some Test List",
		  "icon": "some-icon",
           "folder_id": ` + strconv.Itoa(int(folder.ID)) + `
		}
	  }
	}`,
		},
		{
			name: "data.attributes.icon",
			content: `{
	  "data": {
		"type": "lists",
		"attributes": {
		  "name": "Some Test List",
		  "icon": "some-icon",
           "folder_id": null
		}
	  }
	}`,
		},
	}

	var errorData JsonapiErrors

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := generateRequestWithToken(ts.URL+"/api/v1/lists/", token.Plaintext, "POST", bytes.NewBuffer([]byte(tt.content)))
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
