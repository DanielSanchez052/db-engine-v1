package database

import (
	"db-engine-v1/internal/catalog"
	"db-engine-v1/internal/storage"
	"db-engine-v1/internal/storage/filemanager"
	"db-engine-v1/internal/storage/heapfile"
	"db-engine-v1/internal/storage/pager"
	"db-engine-v1/internal/storage/record"
	"path/filepath"
)

type Database struct {
	fileManager *filemanager.FileManager
	pager       *pager.Pager

	catalog *catalog.CatalogManager

	header *DatabaseHeader
}

func Create(path string) (*Database, error) {
	exists, err := filemanager.FileExists(path)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDatabaseExists
	}

	file, err := filemanager.Open(path)
	if err != nil {
		return nil, err
	}

	fileName := filepath.Base(path)
	filePath := filepath.Dir(path)
	catalogNameFile := filePath + string(filepath.Separator) + "catalog_" + fileName

	catalogManager, err := catalog.Create(catalogNameFile)
	if err != nil {
		return nil, err
	}

	dbHeader := NewDatabaseHeader()

	pager := pager.NewPager(file)

	db := &Database{
		fileManager: file,
		pager:       pager,
		header:      dbHeader,
		catalog:     catalogManager,
	}

	err = db.saveHeader()
	if err != nil {
		file.Close()
		return nil, err
	}

	return db, nil
}

func Open(path string) (*Database, error) {
	file, err := filemanager.Open(path)
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, storage.PageSize)
	err = file.ReadAt(0, buffer)
	if err != nil {
		file.Close()
		return nil, err
	}

	header, err := NewDatabaseHeaderFromBytes(buffer[:storage.DatabaseHeaderSize])
	if err != nil {
		file.Close()
		return nil, err
	}

	if header.MagicNumber != MagicNumber {
		return nil, ErrInvalidDatabase
	}

	pager := pager.NewPager(file)

	fileName := filepath.Base(path)
	filePath := filepath.Dir(path)
	catalogNameFile := filePath + string(filepath.Separator) + "catalog_" + fileName

	catalogManager, err := catalog.Open(catalogNameFile)
	if err != nil {
		return nil, err
	}

	return &Database{
		fileManager: file,
		pager:       pager,
		header:      header,
		catalog:     catalogManager,
	}, nil
}

func (db *Database) Close() error {
	err := db.saveHeader()
	if err != nil {
		return err
	}

	err = db.catalog.Close()
	if err != nil {
		return err
	}

	return db.fileManager.Close()
}

func (db *Database) saveHeader() error {
	dbHeader := db.header

	page0 := make([]byte, storage.PageSize)
	headerBytes, err := dbHeader.Serialize()

	if err != nil {
		return err
	}
	copy(page0, headerBytes)

	//check this later
	err = db.fileManager.WriteAt(0, page0)
	if err != nil {
		return err
	}

	return db.fileManager.Sync()
}

func (db *Database) Pager() *pager.Pager {
	return db.pager
}

func (db *Database) Header() *DatabaseHeader {
	return db.header
}

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

func (db *Database) Insert(tableName string, record record.Record) (*heapfile.RecordID, error) {
	if len(tableName) == 0 {
		return nil, ErrInvalidTableName
	}

	if _, exists := db.catalog.GetTable(tableName); exists {
		return nil, ErrTableAlreadyExists
	}

	heap, err := db.OpenHeapFile(tableName)
	if err != nil {
		return nil, err
	}

	return heap.InsertRecord(record)
}

func (db *Database) GetRecord(tableName string, rid *heapfile.RecordID) (record.Record, error) {
	if len(tableName) == 0 {
		return record.Record{}, ErrInvalidTableName
	}

	if rid == nil {
		return record.Record{}, ErrInvalidRecordID
	}

	heap, err := db.OpenHeapFile(tableName)
	if err != nil {
		return record.Record{}, err
	}

	rec, err := heap.GetRecord(rid)
	if err != nil {
		return record.Record{}, err
	}

	return rec, nil
}
