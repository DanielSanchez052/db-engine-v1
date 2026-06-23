package database

import "db-engine-v1/internal/storage/heapfile"

func (db *Database) OpenHeapFile(tableName string) (*heapfile.HeapFile, error) {
	if len(tableName) == 0 {
		return nil, ErrInvalidTableName
	}

	table, exists := db.catalog.GetTable(tableName)
	if !exists {
		return nil, ErrTableNotFound
	}

	heap, exists := db.catalog.GetHeap(table.HeapName)
	if !exists {
		return nil, ErrTableNotFound
	}

	return heapfile.New(db.Pager(), heap, db.AllocatePage), nil
}
