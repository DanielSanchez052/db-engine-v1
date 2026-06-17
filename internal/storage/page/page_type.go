package page

type PageType uint8

const (
	DataPage PageType = iota
	IndexPage
	CatalogPage
)

func (t PageType) IsValid() bool {
	return t == DataPage || t == IndexPage || t == CatalogPage
}
