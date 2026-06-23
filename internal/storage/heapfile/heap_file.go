package heapfile

import (
	"db-engine-v1/internal/catalog"
	"db-engine-v1/internal/storage/page"
	"db-engine-v1/internal/storage/pager"
)

type HeapFile struct {
	pager        *pager.Pager
	metadata     *catalog.HeapMetadata
	allocatePage func(page.PageType) (*page.Page, error) // TODO: por el momento lo vamos a manejar asi, sin embargo luego se debe mover a otro lado
}

func New(pager *pager.Pager, metadata *catalog.HeapMetadata, allocatePage func(page.PageType) (*page.Page, error)) *HeapFile {
	return &HeapFile{
		pager:        pager,
		metadata:     metadata,
		allocatePage: allocatePage,
	}
}
