package database

import (
	"db-engine-v1/internal/catalog"
	"db-engine-v1/internal/storage/filemanager"
	"db-engine-v1/internal/storage/pager"
)

type Database struct {
	fileManager *filemanager.FileManager
	pager       *pager.Pager

	catalog *catalog.CatalogManager

	header *DatabaseHeader
}

func (db *Database) Pager() *pager.Pager {
	return db.pager
}

func (db *Database) Header() *DatabaseHeader {
	return db.header
}
