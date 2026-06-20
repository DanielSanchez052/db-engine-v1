package catalog

import (
	"encoding/binary"
	"errors"
)

var (
	ErrInvalidTable = errors.New("invalid table")
)

type Table struct {
	Name     string
	HeapName string
	Columns  []Column
}

func NewTableFromBytes(data []byte) (*Table, error) {
	offset := 0

	name, err := readString(data, &offset)
	if err != nil {
		return nil, ErrInvalidTable
	}

	heapName, err := readString(data, &offset)
	if err != nil {
		return nil, ErrInvalidTable
	}

	if offset+ColumnCountSize > len(data) {
		return nil, ErrInvalidTable
	}

	columnCount := binary.LittleEndian.Uint16(data[offset : offset+ColumnCountSize])
	offset += ColumnCountSize

	columns := make([]Column, 0, columnCount)
	for i := uint16(0); i < columnCount; i++ {
		if offset+ColumnLengthSize > len(data) {
			return nil, ErrInvalidTable
		}
		columnLength := binary.LittleEndian.Uint16(data[offset : offset+ColumnLengthSize])
		offset += ColumnLengthSize

		if offset+int(columnLength) > len(data) {
			return nil, ErrInvalidTable
		}
		columnData := data[offset : offset+int(columnLength)]
		column, err := NewColumnFromBytes(columnData)
		if err != nil {
			return nil, ErrInvalidTable
		}

		columns = append(columns, *column)
		offset += int(columnLength)
	}

	return &Table{
		Name:     name,
		HeapName: heapName,
		Columns:  columns,
	}, nil
}

func (t *Table) Serialize() ([]byte, error) {
	buffer := make([]byte, t.Size())
	offset := 0

	if err := writeString(buffer, &offset, t.Name); err != nil {
		return nil, err
	}

	if err := writeString(buffer, &offset, t.HeapName); err != nil {
		return nil, err
	}

	binary.LittleEndian.PutUint16(buffer[offset:offset+ColumnCountSize], uint16(len(t.Columns)))
	offset += ColumnCountSize

	for _, column := range t.Columns {
		columnBytes, err := column.Serialize()
		if err != nil {
			return nil, err
		}

		binary.LittleEndian.PutUint16(buffer[offset:offset+ColumnLengthSize], uint16(len(columnBytes)))
		offset += ColumnLengthSize

		copy(buffer[offset:offset+len(columnBytes)], columnBytes)
		offset += len(columnBytes)
	}

	return buffer, nil
}

func (t *Table) Size() int {
	size := StringLengthSize + len(t.Name) + StringLengthSize + len(t.HeapName) + ColumnCountSize

	for _, column := range t.Columns {
		size += ColumnLengthSize
		size += column.Size()
	}

	return size
}
