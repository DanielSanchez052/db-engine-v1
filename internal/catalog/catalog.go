package catalog

import (
	"db-engine-v1/internal/storage/filemanager"
	"encoding/binary"
)

// CatalogManager manages database schema and metadata
// este por el momento se manejara como archivo binario secuencial, luego mas adelante se implementara un heapfile ya que es mas optimo
type CatalogManager struct {
	fileManager *filemanager.FileManager

	tables map[string]*Table
	heaps  map[string]*HeapMetadata
}

func Create(path string) (*CatalogManager, error) {
	exists, err := filemanager.FileExists(path)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrCatalogAlreadyExists
	}

	file, err := filemanager.Open(path)
	if err != nil {
		return nil, err
	}

	newCatalog := &CatalogManager{
		fileManager: file,
		tables:      make(map[string]*Table),
		heaps:       make(map[string]*HeapMetadata),
	}

	err = newCatalog.save()
	if err != nil {
		return nil, err
	}

	return newCatalog, nil
}

func Open(path string) (*CatalogManager, error) {
	file, err := filemanager.Open(path)
	if err != nil {
		return nil, err
	}

	catalog := &CatalogManager{
		fileManager: file,
		tables:      make(map[string]*Table),
		heaps:       make(map[string]*HeapMetadata),
	}

	err = catalog.load()
	if err != nil {
		file.Close()
		return nil, err
	}

	return catalog, nil
}

func (c *CatalogManager) Flush() error {
	return c.save()
}

func (c *CatalogManager) Close() error {
	err := c.save()
	if err != nil {
		return err
	}
	return c.fileManager.Close()
}

func (c *CatalogManager) save() error {
	buffer := make([]byte, Uint32Size)
	binary.LittleEndian.PutUint32(buffer, uint32(len(c.tables)))

	for _, table := range c.tables {
		tableBytes, err := table.Serialize()
		if err != nil {
			return err
		}

		tableLenBytes := make([]byte, Uint32Size)
		binary.LittleEndian.PutUint32(tableLenBytes, uint32(len(tableBytes)))

		buffer = append(buffer, tableLenBytes...)
		buffer = append(buffer, tableBytes...)
	}

	heapCountBytes := make([]byte, Uint32Size)
	binary.LittleEndian.PutUint32(heapCountBytes, uint32(len(c.heaps)))
	buffer = append(buffer, heapCountBytes...)

	for _, heap := range c.heaps {
		heapBytes, err := heap.Serialize()
		if err != nil {
			return err
		}

		heapLenBytes := make([]byte, Uint32Size)
		binary.LittleEndian.PutUint32(heapLenBytes, uint32(len(heapBytes)))

		buffer = append(buffer, heapLenBytes...)
		buffer = append(buffer, heapBytes...)
	}

	err := c.fileManager.WriteAt(0, buffer)
	if err != nil {
		return err
	}

	err = c.fileManager.Truncate(int64(len(buffer)))
	if err != nil {
		return err
	}

	return c.fileManager.Sync()
}

func (c *CatalogManager) load() error {
	fileSize, err := c.fileManager.Size()
	if err != nil {
		return err
	}

	buffer := make([]byte, fileSize)

	err = c.fileManager.ReadAt(0, buffer)
	if err != nil {
		return err
	}

	if len(buffer) < Uint32Size {
		return ErrInvalidCatalogFormat
	}

	tableCount := binary.LittleEndian.Uint32(buffer[0:Uint32Size])
	buffer = buffer[Uint32Size:]

	for i := 0; i < int(tableCount); i++ {
		if len(buffer) < Uint32Size {
			return ErrInvalidCatalogFormat
		}
		tableLen := binary.LittleEndian.Uint32(buffer[0:Uint32Size])
		buffer = buffer[Uint32Size:]

		if len(buffer) < int(tableLen) {
			return ErrInvalidCatalogFormat
		}
		table, err := NewTableFromBytes(buffer[:tableLen])
		if err != nil {
			return err
		}
		c.tables[table.Name] = table
		buffer = buffer[tableLen:]
	}

	if len(buffer) < Uint32Size {
		return ErrInvalidCatalogFormat
	}

	heapCount := binary.LittleEndian.Uint32(buffer[0:Uint32Size])
	buffer = buffer[Uint32Size:]

	for i := 0; i < int(heapCount); i++ {
		if len(buffer) < Uint32Size {
			return ErrInvalidCatalogFormat
		}
		heapLen := binary.LittleEndian.Uint32(buffer[0:Uint32Size])
		buffer = buffer[Uint32Size:]

		if len(buffer) < int(heapLen) {
			return ErrInvalidCatalogFormat
		}
		heap, err := NewHeapMetadataFromBytes(buffer[:heapLen])
		if err != nil {
			return err
		}
		c.heaps[heap.Name] = heap
		buffer = buffer[heapLen:]
	}

	return nil
}

func (c *CatalogManager) AddTable(table *Table) error {
	c.tables[table.Name] = table
	return nil
}

func (c *CatalogManager) GetTable(name string) (*Table, bool) {
	table, exists := c.tables[name]
	return table, exists
}

func (c *CatalogManager) AddHeap(heap *HeapMetadata) error {
	c.heaps[heap.Name] = heap
	return nil
}

func (c *CatalogManager) GetHeap(name string) (*HeapMetadata, bool) {
	heap, exists := c.heaps[name]
	return heap, exists
}
