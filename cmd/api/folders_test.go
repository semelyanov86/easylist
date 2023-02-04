package main

import (
	"easylist/internal/data"
	"github.com/google/jsonapi"
	"net/http"
	"reflect"
	"strconv"
	"testing"
)

func TestShowDefaultFolder(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	_, token, err := createTestUserWithToken(t, app)
	if err != nil {
		t.Fatal(err)
	}
	defer ts.Close()
	req := generateRequestWithToken(ts.URL+"/api/v1/folders/1", token.Plaintext)
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

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
	if check.Icon != "fa-folder" {
		t.Errorf("want Icon to be fa-folder, got %s", check.Icon)
	}
}

func TestShowNewCreatedFolder(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	user, token, err := createTestUserWithToken(t, app)
	if err != nil {
		t.Fatal(err)
	}
	defer ts.Close()
	folder, err := createTestFolder(app, user.ID, "", 0)
	if err != nil {
		t.Fatal(err)
	}
	req := generateRequestWithToken(ts.URL+"/api/v1/folders/"+strconv.Itoa(int(folder.ID)), token.Plaintext)
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

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

func TestIndexFolders(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	user, token, err := createTestUserWithToken(t, app)
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
	req := generateRequestWithToken(ts.URL+"/api/v1/folders/", token.Plaintext)
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("want %d status code; got %d", http.StatusOK, resp.StatusCode)
	}

	check, err := jsonapi.UnmarshalManyPayload(resp.Body, reflect.TypeOf(new(data.Folder)))
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
