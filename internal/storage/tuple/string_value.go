package tuple

import (
	"db-engine-v1/internal/catalog"
	"encoding/binary"
	"math"
)

type StringValue struct {
	Value string
}

func NewStringValueFromBytes(data []byte) (*StringValue, error) {
	if len(data) < catalog.StringLengthSize {
		return nil, ErrInvalidValue
	}

	length := int(
		binary.LittleEndian.Uint16(
			data[:catalog.StringLengthSize],
		),
	)

	if length != len(data)-catalog.StringLengthSize {
		return nil, ErrInvalidValue
	}

	return &StringValue{
		Value: string(data[catalog.StringLengthSize : catalog.StringLengthSize+length]),
	}, nil
}

func (s StringValue) Type() catalog.DataType {
	return catalog.TypeStringType
}

func (s StringValue) Serialize() ([]byte, error) {
	size := catalog.StringLengthSize + len(s.Value)

	if len(s.Value) > math.MaxUint16 {
		return nil, ErrInvalidValue
	}

	buffer := make([]byte, size)
	offset := 0

	binary.LittleEndian.PutUint16(buffer[offset:offset+catalog.StringLengthSize], uint16(len(s.Value)))
	offset += catalog.StringLengthSize
	copy(buffer[offset:], []byte(s.Value))

	return buffer, nil
}

func (s StringValue) Size() uint16 {
	return uint16(catalog.StringLengthSize + len(s.Value))
}

func (s StringValue) String() string {
	return s.Value
}
