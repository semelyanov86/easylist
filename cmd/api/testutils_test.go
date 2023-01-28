package main

import (
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
			Db:           nil,
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

type testServer struct {
	*httptest.Server
}

func newTestServer(t *testing.T, h http.Handler) *testServer {
	ts := httptest.NewTLSServer(h)
	return &testServer{ts}
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
