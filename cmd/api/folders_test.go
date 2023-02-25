package main

import (
	"bytes"
	"easylist/internal/data"
	"encoding/json"
	"github.com/google/jsonapi"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"testing"
)

func TestShowDefaultFolder(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	_, token, err := createTestUserWithToken(t, app, "")
	if err != nil {
		t.Fatal(err)
	}
	defer ts.Close()
	req := generateRequestWithToken(ts.URL+"/api/v1/folders/1", token.Plaintext, "", nil)
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

	check := new(data.Folder)

	err = jsonapi.UnmarshalPayload(resp.Body, check)

	if err != nil {
		t.Fatal(err)
	}
	if check.ID != 1 {
		t.Errorf("want ID to equal 1, got %d", check.ID)
	}
	if check.Name != "default" {
		t.Errorf("want Name to be Test Name, got %s", check.Name)
	}
	if check.Icon != "mdi-folder" {
		t.Errorf("want Icon to be mdi-folder, got %s", check.Icon)
	}
	if resp.Header.Get("Content-Type") != "application/vnd.api+json" {
		t.Errorf("want Content-Type to be application/vnd.api+json, got %s", resp.Header.Get("Content-Type"))
	}
}

func TestShowNewCreatedFolder(t *testing.T) {
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
	req := generateRequestWithToken(ts.URL+"/api/v1/folders/"+strconv.Itoa(int(folder.ID)), token.Plaintext, "", nil)
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

	check := new(data.Folder)

	err = jsonapi.UnmarshalPayload(resp.Body, check)

	if err != nil {
		t.Fatal(err)
	}
	if check.ID != folder.ID {
		t.Errorf("want ID to equal %d, got %d", folder.ID, check.ID)
	}
	if check.Name != folder.Name {
		t.Errorf("want Name to be %s, got %s", folder.Name, check.Name)
	}
	if check.Icon != folder.Icon {
		t.Errorf("want Icon to be %s, got %s", folder.Icon, check.Icon)
	}
}

func TestFolderNotFoundAccess(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	defer ts.Close()
	user, token, err := createTestUserWithToken(t, app, "")
	if err != nil {
		t.Fatal(err)
	}
	_, err = createTestFolder(app, user.ID, "", 0)

	if err != nil {
		t.Fatal(err)
	}
	tests := []string{
		"adfad", "4", "-2",
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			req := generateRequestWithToken(ts.URL+"/api/v1/folders/"+tt, token.Plaintext, "", nil)
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

			if resp.StatusCode != http.StatusNotFound {
				t.Errorf("want %d status code; got %d", http.StatusNotFound, resp.StatusCode)
			}
		})
	}
}

func TestIndexFolders(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	user, token, err := createTestUserWithToken(t, app, "")
	if err != nil {
		t.Fatal(err)
	}
	defer ts.Close()
	folder1, err := createTestFolder(app, user.ID, "", 0)
	if err != nil {
		t.Fatal(err)
	}
	folder2, err := createTestFolder(app, user.ID, "Second Folder", 3)
	if err != nil {
		t.Fatal(err)
	}
	req := generateRequestWithToken(ts.URL+"/api/v1/folders/", token.Plaintext, "", nil)
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

	check, err := jsonapi.UnmarshalManyPayload(resp.Body, reflect.TypeOf(new(data.Folder)))
	if err != nil {
		t.Fatal(err)
	}
	if len(check) != 3 {
		t.Errorf("want length of folders to equal 3, got %d", len(check))
	}

	if check[0].(*data.Folder).Name != "default" {
		t.Errorf("want first folder name to be default, got %s", check[0].(data.Folder).Name)
	}
	if check[1].(*data.Folder).Name != folder1.Name {
		t.Errorf("want first folder name to be %s, got %s", check[1].(data.Folder).Name, folder1.Name)
	}
	if check[2].(*data.Folder).Name != folder2.Name {
		t.Errorf("want first folder name to be %s, got %s", check[2].(data.Folder).Name, folder2.Name)
	}
}

func TestFolderCreationProcess(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	_, token, err := createTestUserWithToken(t, app, "")
	if err != nil {
		t.Fatal(err)
	}
	defer ts.Close()

	var folderData = []byte(`{
	  "data": {
		"type": "folders",
		"attributes": {
		  "name": "New testing folder",
		  "icon": "mdi-some"
		}
	  }
	}`)

	req := generateRequestWithToken(ts.URL+"/api/v1/folders/", token.Plaintext, "POST", bytes.NewBuffer(folderData))
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

	check := new(data.Folder)

	err = jsonapi.UnmarshalPayload(resp.Body, check)

	if err != nil {
		t.Fatal(err)
	}
	if check.ID < 1 {
		t.Errorf("want correct ID, got %d", check.ID)
	}
	if check.Name != "New testing folder" {
		t.Errorf("want Name to be %s, got %s", "New testing folder", check.Name)
	}
	if check.Icon != "mdi-some" {
		t.Errorf("want Icon to be mdi-some, got %s", check.Icon)
	}
	if check.Order != 1 {
		t.Errorf("want Order to be 1, got %d", check.Order)
	}
	if resp.Header.Get("Location") != "http://127.0.0.1/api/v1/folders/"+strconv.Itoa(int(check.ID)) {
		t.Errorf("want location header with folders value, got %s", resp.Header.Get("Location"))
	}
	if resp.Header.Get("Content-Type") != "application/vnd.api+json" {
		t.Errorf("want Content-Type to be application/vnd.api+json, got %s", resp.Header.Get("Content-Type"))
	}
}

func TestFolderValidation(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	defer ts.Close()
	user, token, err := createTestUserWithToken(t, app, "")
	if err != nil {
		t.Fatal(err)
	}
	_, err = createTestFolder(app, user.ID, "", 0)

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
				"type": "folders",
				"attributes": {
				  "name": "",
				  "icon": "mdi-some"
				}
			  }
			}`,
		},
		{
			name: "data.attributes.icon",
			content: `{
			  "data": {
				"type": "folders",
				"attributes": {
				  "name": "Some folder",
				  "icon": "icicic"
				}
			  }
			}`,
		},
	}

	var errorData JsonapiErrors

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := generateRequestWithToken(ts.URL+"/api/v1/folders/", token.Plaintext, "POST", bytes.NewBuffer([]byte(tt.content)))
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
				t.Errorf("want error title to be %s; got %s", "Validation failed for field", errorData.Errors[0].Title)
			}

		})
	}
}

func TestUpdateFolderProcess(t *testing.T) {
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

	var folderData = []byte(`{
	  "data": {
		"type": "folders",
		"attributes": {
		  "name": "Some new folder name",
		  "icon": "mdi-some-sec",
		"order": 4
		}
	  }
	}`)

	req := generateRequestWithToken(ts.URL+"/api/v1/folders/"+strconv.Itoa(int(folder.ID)), token.Plaintext, "PATCH", bytes.NewBuffer(folderData))
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

	check := new(data.Folder)

	err = jsonapi.UnmarshalPayload(resp.Body, check)

	if err != nil {
		t.Fatal(err)
	}
	if check.ID != folder.ID {
		t.Errorf("want ID to be %d, got %d", folder.ID, check.ID)
	}
	if check.Name != "Some new folder name" {
		t.Errorf("want Name to be %s, got %s", "Some new folder name", check.Name)
	}
	if check.Icon != "mdi-some-sec" {
		t.Errorf("want Icon to be %s, got %s", "mdi-some-sec", check.Icon)
	}
	if check.Order != 4 {
		t.Errorf("want Order to be 4, got %d", check.Order)
	}
	if resp.Header.Get("Content-Type") != "application/vnd.api+json" {
		t.Errorf("want Content-Type to be application/vnd.api+json, got %s", resp.Header.Get("Content-Type"))
	}
}

func TestDeleteFolder(t *testing.T) {
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

	req := generateRequestWithToken(ts.URL+"/api/v1/folders/"+strconv.Itoa(int(folder.ID)), token.Plaintext, "DELETE", nil)
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

}
