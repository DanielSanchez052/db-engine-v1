package tuple

import (
	"db-engine-v1/internal/catalog"
	"encoding/binary"
)

type Tuple struct {
	Values []Value
}

func (t *Tuple) Serialize() ([]byte, error) {
	var result []byte

	for _, value := range t.Values {
		data, err := value.Serialize()
		if err != nil {
			return nil, err
		}
		result = append(result, data...)
	}

	return result, nil
}

func (t *Tuple) Size() uint16 {
	var size uint16 = 0
	for _, value := range t.Values {
		size += value.Size()
	}
	return size
}

func NewTupleFromBytes(data []byte, columns []catalog.Column) (*Tuple, error) {
	if len(columns) == 0 {
		return nil, ErrInvalidTuple
	}

	offset := 0

	var values []Value

	for _, column := range columns {
		size, err := valueSize(data[offset:], column.Type)
		if err != nil {
			return nil, err
		}
		valueData := data[offset : offset+int(size)]

		value, err := valueFromBytes(valueData, column.Type)
		if err != nil {
			return nil, err
		}

		values = append(values, value)
		offset += int(size)
	}

	if offset != len(data) {
		return nil, ErrInvalidTuple
	}

	return &Tuple{
		Values: values,
	}, nil
}

func valueFromBytes(data []byte, dataType catalog.DataType) (Value, error) {
	if !dataType.IsValid() {
		return nil, ErrInvalidDataType
	}

	if len(data) == 0 {
		return nil, ErrInvalidValue
	}

	switch dataType {
	case catalog.TypeInt32Type:
		return NewInt32ValueFromBytes(data)
	case catalog.TypeStringType:
		return NewStringValueFromBytes(data)
	case catalog.TypeBoolType:
		return NewBoolValueFromBytes(data)
	default:
		return nil, ErrInvalidValue
	}
}

func valueSize(data []byte, dataType catalog.DataType) (uint16, error) {
	if !dataType.IsValid() {
		return 0, ErrInvalidDataType
	}

	if len(data) == 0 {
		return 0, ErrInvalidValue
	}

	switch dataType {
	case catalog.TypeInt32Type:
		if len(data) < catalog.Uint32Size {
			return 0, ErrInvalidValue
		}
		return catalog.Uint32Size, nil
	case catalog.TypeStringType:
		if len(data) < catalog.StringLengthSize {
			return 0, ErrInvalidValue
		}

		length := binary.LittleEndian.Uint16(data[:catalog.StringLengthSize])

		if len(data) < int(catalog.StringLengthSize)+int(length) {
			return 0, ErrInvalidValue
		}

		return catalog.StringLengthSize + length, nil
	case catalog.TypeBoolType:
		if len(data) < catalog.BoolSize {
			return 0, ErrInvalidValue
		}
		return catalog.BoolSize, nil
	default:
		return 0, ErrInvalidDataType
	}
}

func (t *Tuple) Validate(columns []catalog.Column) error {
	if len(t.Values) != len(columns) {
		return ErrInvalidTuple
	}

	for i, value := range t.Values {
		if value == nil {
			return ErrInvalidTuple
		}

		if value.Type() != columns[i].Type {
			return ErrInvalidTuple
		}
	}

	return nil
}
