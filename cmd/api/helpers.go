package main

import (
	"encoding/json"
	"errors"
	"maps"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type envelope map[string]any

func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}

func (app *application) writeJson(w http.ResponseWriter, data envelope, status int, headers http.Header) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Uncomment to add extra readability during manual testing:
	// buf := bytes.Buffer{}
	// if err := json.Indent(&buf, b, "", "\t"); err != nil {
	// 	return err
	// }
	// b = buf.Bytes()

	maps.Copy(w.Header(), headers)
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	w.Write(b)

	return nil
}
