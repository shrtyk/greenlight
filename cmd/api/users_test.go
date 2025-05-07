package main

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/shortykevich/greenlight/internal/data"
	"github.com/shortykevich/greenlight/internal/testutils/assertions"
	"github.com/shortykevich/greenlight/internal/testutils/helpers"
)

type userCreateBody struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func TestUsers(t *testing.T) {
	cfg := config{env: "development"}
	models := data.NewMockModels()
	limiter := NewMockLimiter(false)

	app := newApplication(
		withConfig(cfg),
		withLogger(slog.Default()),
		withModels(models),
		withRateLimiter(limiter),
	)

	server := app.routes()

	t.Run("user creation", func(t *testing.T) {
		input := userCreateBody{
			Email:    "shortyk@example.com",
			Name:     "shortyk",
			Password: "pa55word",
		}

		user1Input := helpers.MustJSON(t, input)

		rw := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "/v1/users", user1Input)
		assertions.AssertNoError(t, err)

		server.ServeHTTP(rw, req)

		gotUser := helpers.ReadResp[data.User](t, rw.Result())

		want := envelope{
			"user": data.User{
				ID:        1,
				Email:     "shortyk@example.com",
				Name:      "shortyk",
				CreatedAt: data.MockTimeStamp,
				Activated: false,
			},
		}

		if !reflect.DeepEqual(gotUser["user"], want["user"]) {
			t.Errorf("\ngot : %+v\nwant: %+v", gotUser["user"], want)
		}

		if rw.Code != http.StatusCreated {
			t.Errorf("got: %v, want: %v", rw.Code, http.StatusCreated)
		}
	})
}
