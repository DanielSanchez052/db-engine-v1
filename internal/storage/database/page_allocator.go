package database

import (
	"db-engine-v1/internal/storage/page"
)

func (db *Database) AllocatePage(pageType page.PageType) (*page.Page, error) {

	pageID := db.header.TotalPages

	newPage := page.NewPage(pageID, pageType)

	err := db.pager.SavePage(newPage)
	if err != nil {
		return nil, err
	}

	oldTotalPages := db.header.TotalPages
	db.header.TotalPages++

	err = db.saveHeader()
	if err != nil {
		db.header.TotalPages = oldTotalPages
		return nil, err
	}

	return newPage, nil
}

func (db *Database) FreePage(pageID uint64) error {
	return nil
}
