package database

import "db-engine-v1/internal/catalog"

func (db *Database) CreateTable(tableName string, columns []catalog.Column) error {
	if len(tableName) == 0 {
		return ErrInvalidTableName
	}

	if len(columns) == 0 {
		return ErrInvalidColumns
	}

	heap := &catalog.HeapMetadata{
		Name: tableName + "_heap",
	}

	table := &catalog.Table{
		Name:     tableName,
		HeapName: heap.Name,
		Columns:  columns,
	}

	db.catalog.AddTable(table)
	db.catalog.AddHeap(heap)

	if err := db.catalog.Flush(); err != nil {
		return err
	}

	return nil
}
