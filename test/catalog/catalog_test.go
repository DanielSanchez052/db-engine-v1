package catalog_test

import (
	"testing"

	"db-engine-v1/internal/catalog"
)

func TestCatalogCreate(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/catalog.db"

	cat, err := catalog.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	defer cat.Close()

	if cat == nil {
		t.Fatalf("Create() returned nil")
	}
}

func TestCatalogCreateAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/catalog.db"

	cat, err := catalog.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	cat.Close()

	_, err = catalog.Create(path)
	if err == nil {
		t.Errorf("Create() expected error for existing file, got nil")
	}
}

func TestCatalogOpenNotFound(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/nonexistent.db"

	cat, err := catalog.Open(path)
	if cat != nil {
		cat.Close()
	}
	if err == nil {
		t.Errorf("Open() expected error for non-existent file, got nil")
	}
}

func TestCatalogOpen(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/catalog.db"

	cat, err := catalog.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	cat.Close()

	cat2, err := catalog.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v, want nil", err)
	}
	defer cat2.Close()
}

func TestCatalogAddGetTable(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/catalog.db"

	cat, err := catalog.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer cat.Close()

	table := &catalog.Table{
		Name:     "users",
		HeapName: "users_heap",
		Columns: []catalog.Column{
			{Name: "id", Type: catalog.TypeInt},
			{Name: "name", Type: catalog.TypeString},
		},
	}

	err = cat.AddTable(table)
	if err != nil {
		t.Fatalf("AddTable() error = %v, want nil", err)
	}

	got, exists := cat.GetTable("users")
	if !exists {
		t.Fatalf("GetTable() exists = false, want true")
	}

	if got.Name != "users" {
		t.Errorf("Name = %q, want %q", got.Name, "users")
	}

	if got.HeapName != "users_heap" {
		t.Errorf("HeapName = %q, want %q", got.HeapName, "users_heap")
	}

	if len(got.Columns) != 2 {
		t.Fatalf("len(Columns) = %d, want 2", len(got.Columns))
	}
}

func TestCatalogGetTableNotFound(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/catalog.db"

	cat, err := catalog.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer cat.Close()

	_, exists := cat.GetTable("nonexistent")
	if exists {
		t.Errorf("GetTable() exists = true, want false for nonexistent table")
	}
}

func TestCatalogAddGetHeap(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/catalog.db"

	cat, err := catalog.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer cat.Close()

	heap := &catalog.HeapMetadata{
		Name:    "users_heap",
		PageIDs: []uint64{1, 2, 3},
	}

	err = cat.AddHeap(heap)
	if err != nil {
		t.Fatalf("AddHeap() error = %v, want nil", err)
	}

	got, exists := cat.GetHeap("users_heap")
	if !exists {
		t.Fatalf("GetHeap() exists = false, want true")
	}

	if got.Name != "users_heap" {
		t.Errorf("Name = %q, want %q", got.Name, "users_heap")
	}

	if len(got.PageIDs) != 3 {
		t.Fatalf("len(PageIDs) = %d, want 3", len(got.PageIDs))
	}
}

func TestCatalogGetHeapNotFound(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/catalog.db"

	cat, err := catalog.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer cat.Close()

	_, exists := cat.GetHeap("nonexistent")
	if exists {
		t.Errorf("GetHeap() exists = true, want false")
	}
}

func TestCatalogPersistTablesAndHeaps(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/catalog.db"

	cat, err := catalog.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	table := &catalog.Table{
		Name:     "products",
		HeapName: "products_heap",
		Columns: []catalog.Column{
			{Name: "id", Type: catalog.TypeInt},
			{Name: "name", Type: catalog.TypeString},
			{Name: "price", Type: catalog.TypeInt},
		},
	}
	cat.AddTable(table)

	heap := &catalog.HeapMetadata{
		Name:    "products_heap",
		PageIDs: []uint64{1, 2, 3, 4, 5},
	}
	cat.AddHeap(heap)

	err = cat.Close()
	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	cat2, err := catalog.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer cat2.Close()

	gotTable, exists := cat2.GetTable("products")
	if !exists {
		t.Fatalf("GetTable() exists = false after reopen")
	}

	if gotTable.Name != "products" {
		t.Errorf("Table name = %q, want %q", gotTable.Name, "products")
	}

	if len(gotTable.Columns) != 3 {
		t.Errorf("len(Columns) = %d, want 3", len(gotTable.Columns))
	}

	gotHeap, exists := cat2.GetHeap("products_heap")
	if !exists {
		t.Fatalf("GetHeap() exists = false after reopen")
	}

	if len(gotHeap.PageIDs) != 5 {
		t.Errorf("len(PageIDs) = %d, want 5", len(gotHeap.PageIDs))
	}
}

func TestCatalogMultipleEntities(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/catalog.db"

	cat, err := catalog.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	tables := []*catalog.Table{
		{Name: "users", HeapName: "users_heap", Columns: []catalog.Column{{Name: "id", Type: catalog.TypeInt}}},
		{Name: "orders", HeapName: "orders_heap", Columns: []catalog.Column{{Name: "id", Type: catalog.TypeInt}, {Name: "total", Type: catalog.TypeInt}}},
		{Name: "products", HeapName: "products_heap", Columns: []catalog.Column{{Name: "id", Type: catalog.TypeInt}, {Name: "name", Type: catalog.TypeString}, {Name: "price", Type: catalog.TypeInt}}},
	}

	heaps := []*catalog.HeapMetadata{
		{Name: "users_heap", PageIDs: []uint64{1}},
		{Name: "orders_heap", PageIDs: []uint64{2, 3}},
		{Name: "products_heap", PageIDs: []uint64{4, 5, 6}},
	}

	for _, table := range tables {
		cat.AddTable(table)
	}

	for _, heap := range heaps {
		cat.AddHeap(heap)
	}

	cat.Close()

	cat2, err := catalog.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer cat2.Close()

	for _, table := range tables {
		got, exists := cat2.GetTable(table.Name)
		if !exists {
			t.Errorf("GetTable(%q) exists = false after reopen", table.Name)
			continue
		}
		if len(got.Columns) != len(table.Columns) {
			t.Errorf("Table %q column count = %d, want %d", table.Name, len(got.Columns), len(table.Columns))
		}
	}

	for _, heap := range heaps {
		got, exists := cat2.GetHeap(heap.Name)
		if !exists {
			t.Errorf("GetHeap(%q) exists = false after reopen", heap.Name)
			continue
		}
		if len(got.PageIDs) != len(heap.PageIDs) {
			t.Errorf("Heap %q page count = %d, want %d", heap.Name, len(got.PageIDs), len(heap.PageIDs))
		}
	}
}

func TestCatalogEmpty(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/catalog.db"

	cat, err := catalog.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	cat.Close()

	cat2, err := catalog.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer cat2.Close()

	_, exists := cat2.GetTable("anything")
	if exists {
		t.Errorf("GetTable() exists = true, want false for empty catalog")
	}

	_, exists = cat2.GetHeap("anything")
	if exists {
		t.Errorf("GetHeap() exists = true, want false for empty catalog")
	}
}

func TestCatalogAddTableThenGetHeap(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/catalog.db"

	cat, err := catalog.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer cat.Close()

	table := &catalog.Table{
		Name:     "test",
		HeapName: "test_heap",
		Columns:  []catalog.Column{{Name: "col1", Type: catalog.TypeInt}},
	}
	cat.AddTable(table)

	heap := &catalog.HeapMetadata{
		Name:    "test_heap",
		PageIDs: []uint64{10, 20},
	}
	cat.AddHeap(heap)

	_, exists := cat.GetTable("test")
	if !exists {
		t.Errorf("GetTable() table not found before close, want found")
	}

	_, exists = cat.GetHeap("test_heap")
	if !exists {
		t.Errorf("GetHeap() heap not found before close, want found")
	}
}
