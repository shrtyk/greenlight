package main

import (
	"fmt"
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

	req, err := http.NewRequest(http.MethodGet, "/v1/healthcheck", nil)
	assertNoError(t, err)

	resp := httptest.NewRecorder()

	app.healthcheckHandler(resp, req)

	wantBody := fmt.Sprintf("status: available\nenviroment: %s\nversion: 1.0.0\n", app.config.env)
	gotBody, err := io.ReadAll(resp.Body)
	assertNoError(t, err)
	assertStrs(t, string(gotBody), wantBody)
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
