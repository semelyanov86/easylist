package main

import (
	"database/sql"
	"easylist/internal/data"
	"easylist/internal/jsonlog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func newTestApplication(t *testing.T) *application {
	return &application{
		config: config{
			Port:         4000,
			Env:          "development",
			Registration: false,
			Confirmation: false,
			Domain:       "http://127.0.0.1",
			Statistics:   false,
			Db:           &database{Dsn: os.Getenv("EASYLIST_TEST_DB")},
			Smtp: &smtp{
				Host:     "",
				Port:     25,
				Username: "",
				Password: "",
				Sender:   "",
			},
			Limiter: &limiter{
				Rps:     200,
				Burst:   200,
				Enabled: false,
			},
			Cors: struct {
				TrustedOrigins []string `yaml:"trustedOrigins"`
			}{},
		},
		logger: jsonlog.New(os.Stdout, jsonlog.LevelError),
	}
}

func newTestAppWithDb(t *testing.T) (*application, func()) {
	var app = newTestApplication(t)
	db, teardown := newTestDB(t)
	app.models.Users = data.UserModel{DB: db}
	app.models.Folders = data.FolderModel{DB: db}
	app.models.Tokens = data.TokenModel{DB: db}
	app.models.Permissions = data.PermissionModel{DB: db}
	app.models.Lists = data.ListModel{DB: db}
	app.models.Items = data.ItemModel{DB: db}
	return app, teardown
}

func newTestDB(t *testing.T) (*sql.DB, func()) {
	var dbCred = os.Getenv("EASYLIST_TEST_DB")
	db, err := sql.Open("mysql", dbCred)
	if err != nil {
		t.Fatal(err)
	}

	migrations := [...]string{
		"../../migrations/000001_create_users_table.up.sql",
		"../../migrations/000002_create_tokens_table.up.sql",
		"../../migrations/000003_create_permissions_table.up.sql",
		"../../migrations/000004_create_folders_table.up.sql",
		"../../migrations/000005_create_lists_table.up.sql",
		"../../migrations/000006_create_items_table.up.sql",
		"../../migrations/000008_add_fulltext_search_to_name_column_in_folders_table.up.sql",
		"../../migrations/000009_add_fulltext_search_index_to_name_column_in_lists_table.up.sql",
		"../../migrations/000010_add_fulltext_search_index_to_name_column_in_items_table.up.sql",
	}
	for _, migration := range migrations {
		script, err := os.ReadFile(migration)
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.Exec(string(script))
		if err != nil {
			t.Fatal(err)
		}
	}
	return db, func() {
		migrations := [...]string{
			"../../migrations/000001_create_users_table.down.sql",
			"../../migrations/000002_create_tokens_table.down.sql",
			"../../migrations/000003_create_permissions_table.down.sql",
			"../../migrations/000004_create_folders_table.down.sql",
			"../../migrations/000005_create_lists_table.down.sql",
			"../../migrations/000006_create_items_table.down.sql",
		}
		for _, migration := range migrations {
			script, err := os.ReadFile(migration)
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.Exec(string(script))
			if err != nil {
				t.Fatal(err)
			}
		}
		db.Close()
	}
}

type testServer struct {
	*httptest.Server
}

func newTestServer(t *testing.T, h http.Handler) *testServer {
	ts := httptest.NewTLSServer(h)
	return &testServer{ts}
}

func generateRequestWithToken(url, token string) *http.Request {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.api+json")
	req.Header.Set("Content-Type", "application/vnd.api+json")
	return req
}

/*func (s *testServer) get(t *testing.T, urlPath string) (int, http.Header, []byte) {
	rs, err := s.Client().Get(s.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	return rs.StatusCode, rs.Header, body
}*/
