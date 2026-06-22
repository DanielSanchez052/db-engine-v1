package tuple

import "db-engine-v1/internal/catalog"

type Value interface {
	Type() catalog.DataType

	Serialize() ([]byte, error)

	Size() uint16

	String() string
}
