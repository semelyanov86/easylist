package main

import (
	"bytes"
	"database/sql"
	"easylist/internal/data"
	"encoding/json"
	"github.com/google/jsonapi"
	"io"
	"net/http"
	"reflect"
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}(resp.Body)

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
	if err != nil {
		t.Fatal(err)
	}
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

func TestListCreate(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	user, token, err := createTestUserWithToken(t, app, "")
	if err != nil {
		t.Fatal(err)
	}
	folder, err := createTestFolder(app, user.ID, "", 0)
	if err != nil {
		t.Fatal(err)
	}
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}(resp.Body)

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
	if err != nil {
		t.Fatal(err)
	}
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
				t.Errorf("want error title to be Validation failed for field %s; got %s", tt.name, errorData.Errors[0].Title)
			}

		})
	}
}

func TestListUpdateWithoutPublic(t *testing.T) {
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

	var list = &data.List{UserId: user.ID, FolderId: folder.ID}
	err = createTestList(app, list)

	if err != nil {
		t.Fatal(err)
	}

	var listData = []byte(`{
	  "data": {
			"id": "` + strconv.Itoa(int(list.ID)) + `",
		"type": "lists",
		"attributes": {
			"name": "New name",
			"order": 5,
			"icon": "fa-new",
		  "is_public": false
		}
	  }
	}`)

	req := generateRequestWithToken(ts.URL+"/api/v1/lists/"+strconv.Itoa(int(list.ID)), token.Plaintext, "PATCH", bytes.NewBuffer(listData))
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

	check := new(data.List)

	err = jsonapi.UnmarshalPayload(resp.Body, check)

	if err != nil {
		t.Fatal(err)
	}
	if check.ID != list.ID {
		t.Errorf("want ID to be %d, got %d", folder.ID, check.ID)
	}
	if check.Name != "New name" {
		t.Errorf("want Name to be %s, got %s", "New name", check.Name)
	}
	if check.Icon != "fa-new" {
		t.Errorf("want Icon to be %s, got %s", "fa-new", check.Icon)
	}

	if check.Order != 5 {
		t.Errorf("want Order to be 5, got %d", check.Order)
	}
	if check.Link.Valid {
		t.Errorf("here should be public link empty")
	}
	if resp.Header.Get("Content-Type") != "application/vnd.api+json" {
		t.Errorf("want Content-Type to be application/vnd.api+json, got %s", resp.Header.Get("Content-Type"))
	}
}

func TestMakingListPublic(t *testing.T) {
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

	var list = &data.List{UserId: user.ID, FolderId: folder.ID}
	err = createTestList(app, list)

	if err != nil {
		t.Fatal(err)
	}

	var listData = []byte(`{
	  "data": {
			"id": "` + strconv.Itoa(int(list.ID)) + `",
		"type": "lists",
		"attributes": {
		  "is_public": true
		}
	  }
	}`)

	req := generateRequestWithToken(ts.URL+"/api/v1/lists/"+strconv.Itoa(int(list.ID)), token.Plaintext, "PATCH", bytes.NewBuffer(listData))
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
	check, err := app.models.Lists.Get(list.ID, user.ID)
	if err != nil {
		t.Error(err)
	}

	if !check.Link.Valid {
		t.Errorf("here should be public link NOT empty")
	}
	if resp.Header.Get("Content-Type") != "application/vnd.api+json" {
		t.Errorf("want Content-Type to be application/vnd.api+json, got %s", resp.Header.Get("Content-Type"))
	}
}

func TestShowListWithIncludedFolderAndItems(t *testing.T) {
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
	var item = data.Item{
		UserId: user.ID,
		ListId: list.ID,
	}
	err = createTestItem(app, &item)
	if err != nil {
		t.Fatal(err)
	}

	req := generateRequestWithToken(ts.URL+"/api/v1/lists/"+strconv.Itoa(int(list.ID))+"?include=folder,items", token.Plaintext, "", nil)
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

	check := new(data.List)

	err = jsonapi.UnmarshalPayload(resp.Body, check)

	if err != nil {
		t.Fatal(err)
	}
	if check.ID != list.ID {
		t.Errorf("want ID to equal %d, got %d", list.ID, check.ID)
	}
	if check.Folder == nil {
		t.Errorf("No included folder in response")
	}
	if check.Folder.ID != folder.ID {
		t.Errorf("want folder id to equal %d, got %d", folder.ID, check.Folder.ID)
	}
	if len(check.Items) != 1 {
		t.Errorf("want length of items to be 1, got %d", len(check.Items))
	}
}

func TestIndexAllLists(t *testing.T) {
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
	list.Order = 2
	err = createTestList(app, &list)
	if err != nil {
		t.Fatal(err)
	}
	var list2 = data.List{}
	list2.FolderId = 1
	list2.UserId = user.ID
	list2.Order = 1
	err = createTestList(app, &list2)
	if err != nil {
		t.Fatal(err)
	}
	req := generateRequestWithToken(ts.URL+"/api/v1/lists?sort=-order", token.Plaintext, "", nil)
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

	check, err := jsonapi.UnmarshalManyPayload(resp.Body, reflect.TypeOf(&data.List{}))

	if err != nil {
		t.Fatal(err)
	}
	if check[0].(*data.List).ID != list2.ID {
		t.Errorf("want first element to be with id %d ; got %d", list2.ID, check[0].(*data.List).ID)
	}
	if check[1].(*data.List).ID != list.ID {
		t.Errorf("want second element to be with id %d ; got %d", list.ID, check[1].(*data.List).ID)
	}
}

func TestAllListsFromFolder(t *testing.T) {
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
	list.Order = 2
	err = createTestList(app, &list)
	if err != nil {
		t.Fatal(err)
	}
	var list2 = data.List{}
	list2.FolderId = 1
	list2.UserId = user.ID
	list2.Order = 1
	err = createTestList(app, &list2)
	if err != nil {
		t.Fatal(err)
	}
	req := generateRequestWithToken(ts.URL+"/api/v1/folders/"+strconv.Itoa(int(folder.ID))+"/lists", token.Plaintext, "", nil)
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

	check, err := jsonapi.UnmarshalManyPayload(resp.Body, reflect.TypeOf(&data.List{}))

	if err != nil {
		t.Fatal(err)
	}
	if len(check) > 1 {
		t.Errorf("want length of lists to be 1; got %d", len(check))
	}
	if check[0].(*data.List).ID != list.ID {
		t.Errorf("want first element to be with id %d ; got %d", list.ID, check[0].(*data.List).ID)
	}
}

func TestDeleteList(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	user, token, err := createTestUserWithToken(t, app, "")
	if err != nil {
		t.Fatal(err)
	}
	defer ts.Close()

	var list = data.List{}
	list.FolderId = 1
	list.UserId = user.ID
	list.Order = 1
	err = createTestList(app, &list)
	if err != nil {
		t.Fatal(err)
	}

	req := generateRequestWithToken(ts.URL+"/api/v1/lists/"+strconv.Itoa(int(list.ID)), token.Plaintext, "DELETE", nil)
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

func TestShowPublicListWithItems(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	user, _, err := createTestUserWithToken(t, app, "")
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
	list.Link = data.Link{
		NullString: sql.NullString{String: "111erdde1", Valid: true},
	}
	err = app.models.Lists.Update(&list, 1)
	if err != nil {
		t.Fatal(err)
	}
	var item = data.Item{
		UserId: user.ID,
		ListId: list.ID,
	}
	err = createTestItem(app, &item)
	if err != nil {
		t.Fatal(err)
	}

	req := generateRequestWithToken(ts.URL+"/api/v1/links/111erdde1?include=items", "", "", nil)
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

	if err != nil {
		t.Error(err)
	}
}
