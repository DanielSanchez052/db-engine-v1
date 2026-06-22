package tuple

import "db-engine-v1/internal/catalog"

type BoolValue struct {
	Value bool
}

func NewBoolValueFromBytes(data []byte) (*BoolValue, error) {
	if len(data) != catalog.BoolSize {
		return nil, ErrInvalidValue
	}

	switch data[0] {
	case 0:
		return &BoolValue{Value: false}, nil
	case 1:
		return &BoolValue{Value: true}, nil
	default:
		return nil, ErrInvalidValue
	}
}

func (b BoolValue) Type() catalog.DataType {
	return catalog.TypeBoolType
}

func (b BoolValue) Serialize() ([]byte, error) {
	if b.Value {
		return []byte{1}, nil
	}

	return []byte{0}, nil
}

func (b BoolValue) Size() uint16 {
	return uint16(catalog.BoolSize)
}

func (b BoolValue) String() string {
	if b.Value {
		return "true"
	}
	return "false"
}
