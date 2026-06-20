package catalog

import (
	"encoding/binary"
	"errors"
)

var (
	ErrInvalidValue = errors.New("invalid value")
)

func writeString(buffer []byte, offset *int, value string) error {
	length := len(value)
	if length > maxStringLength {
		return ErrInvalidValue
	}

	binary.LittleEndian.PutUint16(buffer[*offset:*offset+StringLengthSize], uint16(length))
	*offset += StringLengthSize

	copy(buffer[*offset:*offset+length], []byte(value))
	*offset += length

	return nil
}

func readString(data []byte, offset *int) (string, error) {
	if *offset+StringLengthSize > len(data) {
		return "", ErrInvalidValue
	}

	length := binary.LittleEndian.Uint16(
		data[*offset : *offset+StringLengthSize],
	)

	*offset += StringLengthSize

	if int(length) > maxStringLength {
		return "", ErrInvalidValue
	}

	if *offset+int(length) > len(data) {
		return "", ErrInvalidValue
	}

	value := string(
		data[*offset : *offset+int(length)],
	)

	*offset += int(length)

	return value, nil
}

func writeUint64(buffer []byte, offset *int, value uint64) {
	binary.LittleEndian.PutUint64(
		buffer[*offset:*offset+Uint64Size],
		value,
	)

	*offset += Uint64Size
}

func readUint64(data []byte, offset *int) (uint64, error) {
	if *offset+Uint64Size > len(data) {
		return 0, ErrInvalidValue
	}

	value := binary.LittleEndian.Uint64(data[*offset : *offset+Uint64Size])
	*offset += Uint64Size

	return value, nil
}
