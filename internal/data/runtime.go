package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

type Runtime int32

func (r Runtime) MarshalJSON() ([]byte, error) {
	return fmt.Appendf([]byte{}, strconv.Quote("%d mins"), r), nil
}

func (r *Runtime) UnmarshalJSON(b []byte) error {
	unqJsonVal, err := strconv.Unquote(string(b))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	parts := strings.Split(unqJsonVal, " ")
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	i, err := strconv.Atoi(parts[0])
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	*r = Runtime(i)
	return nil
}
