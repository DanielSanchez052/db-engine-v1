package catalog

import (
	"errors"
)

var (
	ErrInvalidColumn = errors.New("invalid column")
)

type Column struct {
	Name string
	Type DataType
}

func NewColumnFromBytes(data []byte) (*Column, error) {
	offset := 0
	name, err := readString(data, &offset)
	if err != nil {
		return nil, err
	}

	if offset+DataTypeSize > len(data) {
		return nil, ErrInvalidColumn
	}

	columnType := DataType(data[offset])

	if !columnType.IsValid() {
		return nil, ErrInvalidColumn
	}

	offset++
	return &Column{
		Name: name,
		Type: columnType,
	}, nil
}

func (c *Column) Serialize() ([]byte, error) {
	if !c.Type.IsValid() {
		return nil, ErrInvalidColumn
	}

	buffer := make([]byte, c.Size())
	offset := 0

	if err := writeString(buffer, &offset, c.Name); err != nil {
		return nil, err
	}

	buffer[offset] = byte(c.Type)

	return buffer, nil
}

func (c *Column) Size() int {
	return StringLengthSize + len(c.Name) + DataTypeSize
}
