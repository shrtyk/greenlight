package validator_test

import (
	"testing"

	"github.com/shortykevich/greenlight/internal/validator"
)

func TestValidator(t *testing.T) {
	t.Run("validations test", func(t *testing.T) {
		cases := map[string]string{
			"right-email": "asd@gmail.com",
			"wrong-email": "!das@gmail.com",
			"password":    "v",
		}

		v := validator.New()
		v.Check(validator.Matches(cases["right-email"], validator.EmailRX), "email", "wrong email")
		if !v.Valid() {
			t.Errorf("values should be valid. Errors: %+v", v.Errors)
		}

		v.Check(validator.Matches(cases["wrong-email"], validator.EmailRX), "email", "wrong email")
		v.Check(len(cases["short-val"]) >= 8, "password", "password length should longer than 8 characters")
		if v.Valid() {
			t.Errorf("values shouldn't be valid at this point. Errors: %+v", v.Errors)
		}
	})

	t.Run("uniqueness test", func(t *testing.T) {
		uniques := []string{"a", "b", "c"}
		nonUniques := []int{1, 2, 2, 3}
		if !validator.Unique(uniques) {
			t.Errorf("got: %v, want: %v", validator.Unique(uniques), true)
		}
		if validator.Unique(nonUniques) {
			t.Errorf("got: %v, want: %v", true, validator.Unique(uniques))
		}
	})
}
