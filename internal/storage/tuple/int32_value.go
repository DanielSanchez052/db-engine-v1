package tuple

import (
	"db-engine-v1/internal/catalog"
	"encoding/binary"
	"strconv"
)

type Int32Value struct {
	Value int32
}

func NewInt32ValueFromBytes(data []byte) (*Int32Value, error) {
	if len(data) != catalog.Uint32Size {
		return nil, ErrInvalidValue
	}

	return &Int32Value{
		Value: int32(binary.LittleEndian.Uint32(data)),
	}, nil
}

func (i Int32Value) Type() catalog.DataType {
	return catalog.TypeInt32Type
}

func (i Int32Value) Serialize() ([]byte, error) {
	buffer := make([]byte, catalog.Uint32Size)

	binary.LittleEndian.PutUint32(buffer, uint32(i.Value))

	return buffer, nil
}

func (i Int32Value) Size() uint16 {
	return uint16(catalog.Uint32Size)
}

func (i Int32Value) String() string {
	return strconv.FormatInt(int64(i.Value), 10)
}
