package tuple

import "errors"

var (
	ErrInvalidValue    = errors.New("invalid value")
	ErrInvalidDataType = errors.New("invalid data type")
	ErrInvalidTuple    = errors.New("invalid tuple")
)
