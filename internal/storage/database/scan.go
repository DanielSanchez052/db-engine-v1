package database

import (
	"db-engine-v1/internal/catalog"
	"db-engine-v1/internal/storage/heapfile"
	"db-engine-v1/internal/storage/tuple"
	"errors"
)

type Scan struct {
    iterator *heapfile.Iterator
    columns  []catalog.Column
}

func (db *Database) Scan(tableName string) (*Scan, error){
	if len(tableName) == 0 {
		return nil, ErrInvalidTableName
	}

	table, exists := db.catalog.GetTable(tableName)
	if !exists {
		return nil, ErrTableNotFound
	}

	heap, err := db.OpenHeapFile(tableName)
	if err != nil {
		return nil, err
	}

	iterator := heapfile.NewIterator(heap)

	return &Scan{
		iterator: iterator,
		columns:  table.Columns,
	}, nil
}


func (s *Scan) Next() (*tuple.Tuple, *heapfile.RecordID, error){
		rec, rid, err := s.iterator.Next()
		
		if errors.Is(err, heapfile.ErrIteratorDone) {
			return nil, nil, ErrScanDone
		}

		if err != nil {
			return nil, nil, err
		}
		
		tupleObj, err := tuple.NewTupleFromBytes(rec, s.columns)
		if err != nil {
			return nil, nil, err
		}

		return tupleObj, rid, nil
}
