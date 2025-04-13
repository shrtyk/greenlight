package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	cfg := config{env: "development"}
	app := newApplication().
		setConfig(cfg).
		setLogger(nil)

	report := healthReport{
		Status:     "available",
		Enviroment: app.config.env,
		Version:    version,
	}

	req, err := http.NewRequest(http.MethodGet, "/v1/healthcheck", nil)
	assertNoError(t, err)

	resp := httptest.NewRecorder()

	app.healthcheckHandler(resp, req)

	wantBody, err := json.Marshal(report)
	assertNoError(t, err)

	gotBody, err := io.ReadAll(resp.Body)
	assertNoError(t, err)
	assertStrs(t, string(gotBody), string(wantBody))
}

func assertNoError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("didn't expect error but got one: %v", err)
	}
}

func assertStrs(t testing.TB, got, want string) {
	if got != want {
		t.Errorf("\ngot:\n%v\nwant:\n%v", string(got), want)
	}
}
