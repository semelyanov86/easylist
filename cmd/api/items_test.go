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

func TestShowCreatedItem(t *testing.T) {
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}(resp.Body)

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

func TestItemCreation(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	defer ts.Close()
	oldItem, token := createItem(app, t)

	var itemData = []byte(`{
	  "data": {
		"type": "items",
		"attributes": {
		  "name": "More products",
		  "description": "For soupe",
				"quantity": 6,
				"quantity_type": "piece",
				"price": 26,
				"is_starred": false,
				"list_id": ` + strconv.Itoa(int(oldItem.ListId)) + `
		}
	  }
	}`)
	req := generateRequestWithToken(ts.URL+"/api/v1/items", token.Plaintext, "POST", bytes.NewBuffer(itemData))
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

	check := new(data.Item)

	err = jsonapi.UnmarshalPayload(resp.Body, check)

	if err != nil {
		t.Fatal(err)
	}
	if check.ID < 1 {
		t.Errorf("want correct ID, got %d", check.ID)
	}
	if check.Name != "More products" {
		t.Errorf("want Name to be %s, got %s", "More products", check.Name)
	}
	if check.Description != "For soupe" {
		t.Errorf("want Description to be For soupe, got %s", check.Description)
	}
	if check.Order != 2 {
		t.Errorf("want Order to be 2, got %d", check.Order)
	}
	if check.ListId != oldItem.ListId {
		t.Errorf("want listid to be %d, got %d", oldItem.ListId, check.ListId)
	}
	if resp.Header.Get("Location") != "http://127.0.0.1/api/v1/items/"+strconv.Itoa(int(check.ID)) {
		t.Errorf("want location header with items value, got %s", resp.Header.Get("Location"))
	}
	if resp.Header.Get("Content-Type") != "application/vnd.api+json" {
		t.Errorf("want Content-Type to be application/vnd.api+json, got %s", resp.Header.Get("Content-Type"))
	}
}

func TestItemValidation(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	oldItem, token := createItem(app, t)

	tests := []struct {
		name    string
		content string
	}{
		{
			name: "data.attributes.name",
			content: `{
	  "data": {
		"type": "items",
		"attributes": {
		  "name": "",
		  "description": "For soupe",
				"quantity": 6,
				"quantity_type": "piece",
				"price": 26,
				"is_starred": false,
				"list_id": ` + strconv.Itoa(int(oldItem.ListId)) + `
		}
	  }
	}`,
		},
		{
			name: "data.attributes.list_id",
			content: `{
	  "data": {
		"type": "items",
		"attributes": {
		  "name": "More products",
		  "description": "For soupe",
				"quantity": 6,
				"quantity_type": "piece",
				"price": 26,
				"is_starred": false
		}
	  }
	}`,
		},
	}

	var errorData JsonapiErrors

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := generateRequestWithToken(ts.URL+"/api/v1/items/", token.Plaintext, "POST", bytes.NewBuffer([]byte(tt.content)))
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

func TestItemUpdate(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	item, token := createItem(app, t)

	var itemData = []byte(`{
	  "data": {
			"id": "` + strconv.Itoa(int(item.ID)) + `",
		"type": "items",
		"attributes": {
		  "name": "More products 2",
		  "description": "For soupe 2",
				"quantity": 3,
				"quantity_type": "piece",
				"price": 22,
				"order": 4,
				"is_starred": false
		}
	  }
	}`)

	req := generateRequestWithToken(ts.URL+"/api/v1/items/"+strconv.Itoa(int(item.ID)), token.Plaintext, "PATCH", bytes.NewBuffer(itemData))
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

	check := new(data.Item)

	err = jsonapi.UnmarshalPayload(resp.Body, check)

	if err != nil {
		t.Fatal(err)
	}
	if check.ID != item.ID {
		t.Errorf("want ID to be %d, got %d", item.ID, check.ID)
	}
	if check.Name != "More products 2" {
		t.Errorf("want Name to be %s, got %s", "More products 2", check.Name)
	}
	if check.Description != "For soupe 2" {
		t.Errorf("want Description to be %s, got %s", "For soupe 2", check.Description)
	}

	if check.Order != 4 {
		t.Errorf("want Order to be 4, got %d", check.Order)
	}
	if check.Price != 22 {
		t.Errorf("want Price to be 22, got %f", check.Price)
	}
	if check.Quantity != 3 {
		t.Errorf("want Quantity to be 3, got %d", check.Quantity)
	}
	if resp.Header.Get("Content-Type") != "application/vnd.api+json" {
		t.Errorf("want Content-Type to be application/vnd.api+json, got %s", resp.Header.Get("Content-Type"))
	}
}

func TestUpdateOnlyItemOrder(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	item, token := createItem(app, t)

	var itemData = []byte(`{
	  "data": {
			"id": "` + strconv.Itoa(int(item.ID)) + `",
		"type": "items",
		"attributes": {
				"order": 8
		}
	  }
	}`)

	req := generateRequestWithToken(ts.URL+"/api/v1/items/"+strconv.Itoa(int(item.ID)), token.Plaintext, "PATCH", bytes.NewBuffer(itemData))
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

	check := new(data.Item)

	err = jsonapi.UnmarshalPayload(resp.Body, check)

	if err != nil {
		t.Fatal(err)
	}
	if check.ID != item.ID {
		t.Errorf("want ID to be %d, got %d", item.ID, check.ID)
	}
	if check.Name != item.Name {
		t.Errorf("want Name to be %s, got %s", item.Name, check.Name)
	}
	if check.Description != item.Description {
		t.Errorf("want Description to be %s, got %s", item.Description, check.Description)
	}

	if check.Order != 8 {
		t.Errorf("want Order to be 8, got %d", check.Order)
	}
	if check.Price != item.Price {
		t.Errorf("want Price to be %f, got %f", item.Price, check.Price)
	}
	if check.Quantity != item.Quantity {
		t.Errorf("want Quantity to be %d, got %d", item.Quantity, check.Quantity)
	}
	if resp.Header.Get("Content-Type") != "application/vnd.api+json" {
		t.Errorf("want Content-Type to be application/vnd.api+json, got %s", resp.Header.Get("Content-Type"))
	}
}

func TestShowItemWithIncludedList(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	item, token := createItem(app, t)

	req := generateRequestWithToken(ts.URL+"/api/v1/items/"+strconv.Itoa(int(item.ID))+"?include=list", token.Plaintext, "", nil)
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

	check := new(data.Item)

	err = jsonapi.UnmarshalPayload(resp.Body, check)

	if err != nil {
		t.Fatal(err)
	}
	if check.ID != item.ID {
		t.Errorf("want ID to equal %d, got %d", item.ID, check.ID)
	}
	if check.List == nil {
		t.Error("there is no included list in response")
	}
	if check.List.ID != item.ListId {
		t.Errorf("want List ID to equal %d, got %d", item.ListId, check.List.ID)
	}
}

func TestDeleteItem(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	item, token := createItem(app, t)
	req := generateRequestWithToken(ts.URL+"/api/v1/items/"+strconv.Itoa(int(item.ID)), token.Plaintext, "DELETE", nil)
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

func TestIndexOfAllItems(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	item, token := createItem(app, t)
	req := generateRequestWithToken(ts.URL+"/api/v1/items", token.Plaintext, "", nil)
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

	check, err := jsonapi.UnmarshalManyPayload(resp.Body, reflect.TypeOf(&data.Item{}))
	if err != nil {
		t.Fatal(err)
	}
	if check[0].(*data.Item).ID != item.ID {
		t.Errorf("want first element to be with id %d ; got %d", item.ID, check[0].(*data.Item).ID)
	}
}

func TestGettingItemsFromList(t *testing.T) {
	app, teardown := newTestAppWithDb(t)
	defer teardown()

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	item, token := createItem(app, t)
	req := generateRequestWithToken(ts.URL+"/api/v1/lists/"+strconv.Itoa(int(item.ListId))+"/items", token.Plaintext, "", nil)
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

	check, err := jsonapi.UnmarshalManyPayload(resp.Body, reflect.TypeOf(&data.Item{}))
	if err != nil {
		t.Fatal(err)
	}
	if check[0].(*data.Item).ID != item.ID {
		t.Errorf("want first element to be with id %d ; got %d", item.ID, check[0].(*data.Item).ID)
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
