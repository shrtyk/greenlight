package data

import (
	"fmt"
	"strconv"
)

type Runtime int32

func (r Runtime) MarshalJSON() ([]byte, error) {
	return fmt.Appendf([]byte{}, strconv.Quote("%d mins"), r), nil
}
