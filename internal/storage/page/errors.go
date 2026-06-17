package page

import "errors"

var (
	ErrPageHeaderSizeMismatch  = errors.New("page header data size mismatch")
	ErrInvalidPageType         = errors.New("invalid page type")
	ErrInvalidFreeSpaceOffset  = errors.New("free space offset exceeds page size")
	ErrFreeSpaceOffsetInHeader = errors.New("free space offset points within header")
	ErrPageSizeMismatch        = errors.New("page size mismatch")
)
