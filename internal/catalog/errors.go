package catalog

import "errors"

var (
	ErrCatalogAlreadyExists = errors.New("catalog already exists")
	ErrInvalidCatalogFormat = errors.New("invalid catalog format")
)
