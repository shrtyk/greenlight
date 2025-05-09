package helpers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func MustJSON(t *testing.T, v any) *bytes.Reader {
	t.Helper()

	if v == nil {
		return bytes.NewReader([]byte{})
	}

	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal JSON: %v", err)
	}
	return bytes.NewReader(b)
}

func ReadResp[T any](t *testing.T, resp *http.Response) map[string]T {
	t.Helper()

	var tr map[string]T
	err := json.NewDecoder(resp.Body).Decode(&tr)
	if err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	return tr
}
