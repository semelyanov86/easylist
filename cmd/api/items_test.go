package main

import (
	"easylist/internal/data"
	"github.com/google/jsonapi"
	"net/http"
	"strconv"
	"testing"
)

func TestShowCreatedList(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	item, token := createItem(app, t)

	req := generateRequestWithToken(ts.URL+"/api/v1/items/"+strconv.Itoa(int(item.ID)), token.Plaintext, "", nil)
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("want %d status code; got %d", http.StatusOK, resp.StatusCode)
	}

	check := new(data.Item)

	err = jsonapi.UnmarshalPayload(resp.Body, check)

	if err != nil {
		t.Fatal(err)
	}
	if check.ID != item.ID {
		t.Errorf("want ID to equal %d, got %d", item.ID, check.ID)
	}
	if check.Name != item.Name {
		t.Errorf("want Name to be %s, got %s", item.Name, check.Name)
	}
	if check.Description != item.Description {
		t.Errorf("want Description to be %s, got %s", item.Description, check.Description)
	}
}

func TestItemNotFoundAccess(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	_, token := createItem(app, t)

	tests := []string{
		"adfad", "4", "-2",
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			req := generateRequestWithToken(ts.URL+"/api/v1/items/"+tt, token.Plaintext, "", nil)
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

func createItem(app *application, t *testing.T) (data.Item, *data.Token) {
	user, token, err := createTestUserWithToken(t, app, "")
	if err != nil {
		t.Fatal(err)
	}
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

	var item = data.Item{}
	item.ListId = list.ID
	item.UserId = user.ID
	err = createTestItem(app, &item)
	if err != nil {
		t.Fatal(err)
	}
	return item, token
}
