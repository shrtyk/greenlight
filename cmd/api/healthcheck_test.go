package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shrtyk/greenlight/internal/testutils/assertions"
)

func TestHealthCheck(t *testing.T) {
	cfg := config{env: "development"}
	app := newApplication(
		withConfig(cfg),
		withLogger(nil),
	)

	req, err := http.NewRequest(http.MethodGet, "/v1/healthcheck", nil)
	assertions.AssertNoError(t, err)

	resp := httptest.NewRecorder()

	app.healthcheckHandler(resp, req)

	report := envelope{
		"status": "available",
		"system_info": map[string]string{
			"enviroment": app.config.env,
			"version":    app.version,
		},
	}

	wantBody, err := json.Marshal(report)
	assertions.AssertNoError(t, err)

	gotBody, err := io.ReadAll(resp.Body)
	assertions.AssertNoError(t, err)
	assertions.AssertStrings(t, string(gotBody), string(wantBody))
}
