package pager

import (
	"db-engine-v1/internal/storage"
	"db-engine-v1/internal/storage/filemanager"
	"db-engine-v1/internal/storage/page"
)

type Pager struct {
	fileManager *filemanager.FileManager
}

func NewPager(fm *filemanager.FileManager) *Pager {
	return &Pager{
		fileManager: fm,
	}
}

func (p *Pager) pageOffset(pageID uint64) int64 {
	return int64(pageID) * storage.PageSize
}

func (p *Pager) LoadPage(pageID uint64) (*page.Page, error) {
	offset := p.pageOffset(pageID)

	buffer := make([]byte, storage.PageSize)
	err := p.fileManager.ReadAt(offset, buffer)
	if err != nil {
		return nil, err
	}

	loadedPage, err := page.NewPageFromBytes(buffer)
	if err != nil {
		return nil, err
	}

	return loadedPage, nil
}

func (p *Pager) SavePage(page *page.Page) error {
	data, err := page.Serialize()
	if err != nil {
		return err
	}

	offset := p.pageOffset(page.Header.PageID)

	err = p.fileManager.WriteAt(offset, data)
	if err != nil {
		return err
	}

	return p.fileManager.Sync()
}
