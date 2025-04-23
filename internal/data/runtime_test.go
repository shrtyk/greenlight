package data_test

import (
	"strconv"
	"testing"

	"github.com/shortykevich/greenlight/internal/data"
)

func TestRuntimeType(t *testing.T) {
	quotedStr := strconv.Quote("120 mins")

	t.Run("marshalling test", func(t *testing.T) {
		tVal := data.Runtime(120)

		got, err := tVal.MarshalJSON()
		if err != nil {
			t.Fatalf("didn't expect error but got: %v", err)
		}

		want := quotedStr
		if string(got) != want {
			t.Errorf("got %s, wanted %s", string(got), want)
		}
	})

	t.Run("unmarshalling test", func(t *testing.T) {
		var got data.Runtime

		tVal := quotedStr
		if err := got.UnmarshalJSON([]byte(tVal)); err != nil {
			t.Fatalf("didn't expect error but got: %v", err)
		}

		want := data.Runtime(120)
		if got != want {
			t.Errorf("got %v, wanted %v", got, want)
		}
	})
}
