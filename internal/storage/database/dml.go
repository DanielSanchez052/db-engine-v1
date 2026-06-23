package database

import (
	"db-engine-v1/internal/storage/heapfile"
	"db-engine-v1/internal/storage/record"
	"db-engine-v1/internal/storage/tuple"
)

func (db *Database) Insert(tableName string, data *tuple.Tuple) (*heapfile.RecordID, error) {
	if len(tableName) == 0 {
		return nil, ErrInvalidTableName
	}

	table, exists := db.catalog.GetTable(tableName)
	if !exists {
		return nil, ErrTableNotFound
	}

	if err := data.Validate(table.Columns); err != nil {
		return nil, err
	}

	heap, err := db.OpenHeapFile(tableName)
	if err != nil {
		return nil, err
	}

	recordBytes, err := data.Serialize()
	if err != nil {
		return nil, err
	}

	var recordObj record.Record = recordBytes

	rid, metadataChanged, err := heap.InsertRecord(recordObj)
	if err != nil {
		return nil, err
	}

	if metadataChanged {
		if err := db.catalog.Flush(); err != nil {
			return nil, err
		}
	}

	return rid, nil
}

func (db *Database) GetTuple(tableName string, rid *heapfile.RecordID) (tuple.Tuple, error) {
	if len(tableName) == 0 {
		return tuple.Tuple{}, ErrInvalidTableName
	}

	if rid == nil {
		return tuple.Tuple{}, ErrInvalidRecordID
	}

	table, exists := db.catalog.GetTable(tableName)
	if !exists {
		return tuple.Tuple{}, ErrTableNotFound
	}

	heap, err := db.OpenHeapFile(tableName)
	if err != nil {
		return tuple.Tuple{}, err
	}

	rec, err := heap.GetRecord(rid)
	if err != nil {
		return tuple.Tuple{}, err
	}

	recordTuple, err := tuple.NewTupleFromBytes(rec, table.Columns)
	if err != nil {
		return tuple.Tuple{}, err
	}

	return *recordTuple, nil
}
