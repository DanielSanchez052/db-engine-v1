package database

import (
	"db-engine-v1/internal/catalog"
	"db-engine-v1/internal/storage"
	"db-engine-v1/internal/storage/filemanager"
	"db-engine-v1/internal/storage/pager"
	"path/filepath"
)

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
