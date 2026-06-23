package storage_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"testing"

	"db-engine-v1/internal/catalog"
	"db-engine-v1/internal/storage"
	"db-engine-v1/internal/storage/database"
	"db-engine-v1/internal/storage/heapfile"
	"db-engine-v1/internal/storage/page"
	"db-engine-v1/internal/storage/record"
	"db-engine-v1/internal/storage/tuple"
)

func TestNewDatabaseHeader(t *testing.T) {
	h := database.NewDatabaseHeader()

	wantMagic := [4]byte{'M', 'N', 'D', 'B'}
	if h.MagicNumber != wantMagic {
		t.Errorf("MagicNumber = %v, want %v", h.MagicNumber, wantMagic)
	}

	if h.Version != 1 {
		t.Errorf("Version = %d, want %d", h.Version, 1)
	}

	if h.PageSize != storage.PageSize {
		t.Errorf("PageSize = %d, want %d", h.PageSize, storage.PageSize)
	}

	if h.TotalPages != 1 {
		t.Errorf("TotalPages = %d, want %d", h.TotalPages, 1)
	}

	if h.FreePageHead != 0 {
		t.Errorf("FreePageHead = %d, want %d", h.FreePageHead, 0)
	}

	wantReserved := [40]byte{}
	if h.Reserved != wantReserved {
		t.Errorf("Reserved = %v, want %v", h.Reserved, wantReserved)
	}
}

func TestDatabaseHeader_Serialize(t *testing.T) {
	h := database.NewDatabaseHeader()
	h.Version = 2
	h.TotalPages = 100
	h.FreePageHead = 42
	copy(h.Reserved[:], []byte("hello"))

	buf, err := h.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v, want nil", err)
	}

	if len(buf) != storage.DatabaseHeaderSize {
		t.Fatalf("Serialize() len = %d, want %d", len(buf), storage.DatabaseHeaderSize)
	}

	if !bytes.Equal(buf[database.MagicNumberOffset:database.MagicNumberOffset+database.MagicNumberSize], h.MagicNumber[:]) {
		t.Errorf("MagicNumber bytes = %v, want %v", buf[database.MagicNumberOffset:database.MagicNumberOffset+database.MagicNumberSize], h.MagicNumber[:])
	}

	if got := binary.LittleEndian.Uint16(buf[database.VersionOffset : database.VersionOffset+database.VersionSize]); got != h.Version {
		t.Errorf("Version bytes = %d, want %d", got, h.Version)
	}

	if got := binary.LittleEndian.Uint16(buf[database.PageSizeOffset : database.PageSizeOffset+database.PageSizeSize]); got != h.PageSize {
		t.Errorf("PageSize bytes = %d, want %d", got, h.PageSize)
	}

	if got := binary.LittleEndian.Uint64(buf[database.TotalPagesOffset : database.TotalPagesOffset+database.TotalPagesSize]); got != h.TotalPages {
		t.Errorf("TotalPages bytes = %d, want %d", got, h.TotalPages)
	}

	if got := binary.LittleEndian.Uint64(buf[database.FreePageHeadOffset : database.FreePageHeadOffset+database.FreePageHeadSize]); got != h.FreePageHead {
		t.Errorf("FreePageHead bytes = %d, want %d", got, h.FreePageHead)
	}

	if !bytes.Equal(buf[database.ReservedOffset:database.ReservedOffset+database.ReservedSize], h.Reserved[:]) {
		t.Errorf("Reserved bytes = %v, want %v", buf[database.ReservedOffset:database.ReservedOffset+database.ReservedSize], h.Reserved[:])
	}
}

func TestNewDatabaseHeaderFromBytes(t *testing.T) {
	t.Run("size mismatch", func(t *testing.T) {
		cases := []struct {
			name string
			size int
		}{
			{"too short", storage.DatabaseHeaderSize - 1},
			{"too long", storage.DatabaseHeaderSize + 1},
			{"empty", 0},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				data := make([]byte, tc.size)
				h, err := database.NewDatabaseHeaderFromBytes(data)
				if h != nil {
					t.Errorf("NewDatabaseHeaderFromBytes() header = %v, want nil", h)
				}
				if !errors.Is(err, database.ErrHeaderSizeMismatch) {
					t.Errorf("NewDatabaseHeaderFromBytes() error = %v, want %v", err, database.ErrHeaderSizeMismatch)
				}
			})
		}
	})

	t.Run("valid data", func(t *testing.T) {
		original := database.NewDatabaseHeader()
		original.Version = 7
		original.TotalPages = 1234
		original.FreePageHead = 99
		copy(original.Reserved[:], []byte("reserved-data"))

		buf, err := original.Serialize()
		if err != nil {
			t.Fatalf("Serialize() error = %v, want nil", err)
		}

		got, err := database.NewDatabaseHeaderFromBytes(buf)
		if err != nil {
			t.Fatalf("NewDatabaseHeaderFromBytes() error = %v, want nil", err)
		}

		if got.MagicNumber != original.MagicNumber {
			t.Errorf("MagicNumber = %v, want %v", got.MagicNumber, original.MagicNumber)
		}
		if got.Version != original.Version {
			t.Errorf("Version = %d, want %d", got.Version, original.Version)
		}
		if got.PageSize != original.PageSize {
			t.Errorf("PageSize = %d, want %d", got.PageSize, original.PageSize)
		}
		if got.TotalPages != original.TotalPages {
			t.Errorf("TotalPages = %d, want %d", got.TotalPages, original.TotalPages)
		}
		if got.FreePageHead != original.FreePageHead {
			t.Errorf("FreePageHead = %d, want %d", got.FreePageHead, original.FreePageHead)
		}
		if got.Reserved != original.Reserved {
			t.Errorf("Reserved = %v, want %v", got.Reserved, original.Reserved)
		}
	})
}

func TestDatabaseHeader_RoundTrip(t *testing.T) {
	original := database.NewDatabaseHeader()

	buf, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v, want nil", err)
	}

	got, err := database.NewDatabaseHeaderFromBytes(buf)
	if err != nil {
		t.Fatalf("NewDatabaseHeaderFromBytes() error = %v, want nil", err)
	}

	if *got != *original {
		t.Errorf("round trip mismatch: got %+v, want %+v", got, original)
	}
}

func TestDatabaseHeader_Validate(t *testing.T) {
	t.Run("valid header", func(t *testing.T) {
		h := database.NewDatabaseHeader()
		if err := h.Validate(); err != nil {
			t.Errorf("Validate() error = %v, want nil", err)
		}
	})

	t.Run("invalid magic number", func(t *testing.T) {
		h := database.NewDatabaseHeader()
		h.MagicNumber = [4]byte{'X', 'X', 'X', 'X'}

		err := h.Validate()
		if !errors.Is(err, database.ErrInvalidMagicNumber) {
			t.Errorf("Validate() error = %v, want %v", err, database.ErrInvalidMagicNumber)
		}
	})

	t.Run("invalid page size", func(t *testing.T) {
		h := database.NewDatabaseHeader()
		h.PageSize = storage.PageSize + 1

		err := h.Validate()
		if !errors.Is(err, database.ErrInvalidPageSize) {
			t.Errorf("Validate() error = %v, want %v", err, database.ErrInvalidPageSize)
		}
	})

	t.Run("invalid magic number checked before page size", func(t *testing.T) {
		h := database.NewDatabaseHeader()
		h.MagicNumber = [4]byte{'X', 'X', 'X', 'X'}
		h.PageSize = storage.PageSize + 1

		err := h.Validate()
		if !errors.Is(err, database.ErrInvalidMagicNumber) {
			t.Errorf("Validate() error = %v, want %v", err, database.ErrInvalidMagicNumber)
		}
	})
}

func TestDatabase_Create(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	db, err := database.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	defer db.Close()

	if db == nil {
		t.Fatalf("Create() returned nil database")
	}

	if db.Header() == nil {
		t.Fatalf("Create() returned database with nil Header")
	}

	if db.Header().MagicNumber != database.MagicNumber {
		t.Errorf("MagicNumber = %v, want %v", db.Header().MagicNumber, database.MagicNumber)
	}

	if db.Header().Version != database.Version {
		t.Errorf("Version = %d, want %d", db.Header().Version, database.Version)
	}

	if db.Header().PageSize != storage.PageSize {
		t.Errorf("PageSize = %d, want %d", db.Header().PageSize, storage.PageSize)
	}

	if db.Header().TotalPages != 1 {
		t.Errorf("TotalPages = %d, want %d", db.Header().TotalPages, 1)
	}

	// Verify file was created
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Database file was not created at %s", path)
	}
}

func TestDatabase_Open(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	// First create a database
	db, err := database.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	db.Close()

	// Now open it
	db, err = database.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v, want nil", err)
	}
	defer db.Close()

	if db == nil {
		t.Fatalf("Open() returned nil database")
	}

	if db.Header() == nil {
		t.Fatalf("Open() returned database with nil Header")
	}

	if db.Header().MagicNumber != database.MagicNumber {
		t.Errorf("MagicNumber = %v, want %v", db.Header().MagicNumber, database.MagicNumber)
	}

	if db.Header().Version != database.Version {
		t.Errorf("Version = %d, want %d", db.Header().Version, database.Version)
	}

	if db.Header().PageSize != storage.PageSize {
		t.Errorf("PageSize = %d, want %d", db.Header().PageSize, storage.PageSize)
	}

	if db.Header().TotalPages != 1 {
		t.Errorf("TotalPages = %d, want %d", db.Header().TotalPages, 1)
	}
}

func TestDatabase_Open_NotFound(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/nonexistent.db"

	_, err := database.Open(path)
	if err == nil {
		t.Errorf("Open() expected error for nonexistent file, got nil")
	}
}

func TestDatabase_Open_InvalidHeader(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/invalid.db"

	// Create a file with invalid magic number (full page)
	invalidData := make([]byte, storage.PageSize)
	copy(invalidData, []byte("XXXX")) // Invalid magic number

	err := os.WriteFile(path, invalidData, 0644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err = database.Open(path)
	if err == nil {
		t.Errorf("Open() expected error for invalid header, got nil")
	}
	if !errors.Is(err, database.ErrInvalidMagicNumber) {
		t.Errorf("Open() error = %v, want %v", err, database.ErrInvalidMagicNumber)
	}
}

func TestDatabase_Close_SavesHeader(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	db, err := database.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}

	// Modify header
	h := db.Header()
	h.TotalPages = 42
	h.FreePageHead = 7

	err = db.Close()
	if err != nil {
		t.Fatalf("Close() error = %v, want nil", err)
	}

	// Reopen and verify changes persisted
	db2, err := database.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v, want nil", err)
	}
	defer db2.Close()

	if db2.Header().TotalPages != 42 {
		t.Errorf("TotalPages after reopen = %d, want 42", db2.Header().TotalPages)
	}

	if db2.Header().FreePageHead != 7 {
		t.Errorf("FreePageHead after reopen = %d, want 7", db2.Header().FreePageHead)
	}
}

func TestAllocateFirstPage(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	db, err := database.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}

	// Page 0 is metadata, first allocation should be page 1
	p, err := db.AllocatePage(page.DataPage)
	if err != nil {
		t.Fatalf("AllocatePage() error = %v, want nil", err)
	}

	if p.Header.PageID != 1 {
		t.Errorf("Allocated page ID = %d, want 1 (page 0 is metadata)", p.Header.PageID)
	}

	if p.Header.PageType != page.DataPage {
		t.Errorf("PageType = %v, want %v", p.Header.PageType, page.DataPage)
	}

	db.Close()

	// Reopen and verify TotalPages persisted
	db2, err := database.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v, want nil", err)
	}
	defer db2.Close()

	if db2.Header().TotalPages != 2 {
		t.Errorf("TotalPages after reopen = %d, want 2", db2.Header().TotalPages)
	}
}

func TestAllocateMultiplePages(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	db, err := database.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}

	var pageIDs []uint64
	for i := 0; i < 5; i++ {
		p, err := db.AllocatePage(page.DataPage)
		if err != nil {
			t.Fatalf("AllocatePage(%d) error = %v, want nil", i, err)
		}
		pageIDs = append(pageIDs, p.Header.PageID)
	}

	// Verify sequential allocation starting from 1
	for i, id := range pageIDs {
		expected := uint64(i + 1)
		if id != expected {
			t.Errorf("Page %d ID = %d, want %d", i, id, expected)
		}
	}

	db.Close()

	// Reopen and verify all pages persisted
	db2, err := database.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v, want nil", err)
	}
	defer db2.Close()

	if db2.Header().TotalPages != 6 { // 1 metadata + 5 data pages
		t.Errorf("TotalPages after reopen = %d, want 6", db2.Header().TotalPages)
	}

	// Verify each page can be loaded
	for _, id := range pageIDs {
		loadedPage, err := db2.Pager().LoadPage(id)
		if err != nil {
			t.Fatalf("LoadPage(%d) error = %v, want nil", id, err)
		}
		if loadedPage.Header.PageID != id {
			t.Errorf("Loaded page ID = %d, want %d", loadedPage.Header.PageID, id)
		}
		if loadedPage.Header.PageType != page.DataPage {
			t.Errorf("Loaded page type = %v, want %v", loadedPage.Header.PageType, page.DataPage)
		}
	}
}

func TestAllocateDifferentPageTypes(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	db, err := database.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}

	dataPage, err := db.AllocatePage(page.DataPage)
	if err != nil {
		t.Fatalf("AllocatePage(DataPage) error = %v", err)
	}
	if dataPage.Header.PageType != page.DataPage {
		t.Errorf("DataPage type = %v, want %v", dataPage.Header.PageType, page.DataPage)
	}

	indexPage, err := db.AllocatePage(page.IndexPage)
	if err != nil {
		t.Fatalf("AllocatePage(IndexPage) error = %v", err)
	}
	if indexPage.Header.PageType != page.IndexPage {
		t.Errorf("IndexPage type = %v, want %v", indexPage.Header.PageType, page.IndexPage)
	}

	db.Close()
}

func TestCreateTable(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	db, err := database.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer db.Close()

	columns := []catalog.Column{
		{Name: "id", Type: catalog.TypeInt32Type},
		{Name: "name", Type: catalog.TypeStringType},
	}

	err = db.CreateTable("users", columns)
	if err != nil {
		t.Fatalf("CreateTable() error = %v, want nil", err)
	}

	// Verify by opening the heap file and inserting/reading a record
	hf, err := db.OpenHeapFile("users")
	if err != nil {
		t.Fatalf("OpenHeapFile() error = %v, want nil", err)
	}

	rec := record.Record([]byte("test data"))
	rid, _, err := hf.InsertRecord(rec)
	if err != nil {
		t.Fatalf("InsertRecord() error = %v, want nil", err)
	}

	got, err := hf.GetRecord(rid)
	if err != nil {
		t.Fatalf("GetRecord() error = %v, want nil", err)
	}
	if string(got) != "test data" {
		t.Errorf("GetRecord() = %q, want %q", string(got), "test data")
	}
}

func TestCreateTableEmptyName(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	db, err := database.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer db.Close()

	err = db.CreateTable("", []catalog.Column{{Name: "id", Type: catalog.TypeInt32Type}})
	if !errors.Is(err, database.ErrInvalidTableName) {
		t.Errorf("CreateTable() error = %v, want %v", err, database.ErrInvalidTableName)
	}
}

func TestCreateTableEmptyColumns(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	db, err := database.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer db.Close()

	err = db.CreateTable("users", []catalog.Column{})
	if !errors.Is(err, database.ErrInvalidColumns) {
		t.Errorf("CreateTable() error = %v, want %v", err, database.ErrInvalidColumns)
	}
}

func TestCreateTablePersistsAfterReopen(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	db, err := database.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	columns := []catalog.Column{
		{Name: "name", Type: catalog.TypeStringType},
	}
	err = db.CreateTable("users", columns)
	if err != nil {
		t.Fatalf("CreateTable() error = %v", err)
	}

	tup := tuple.NewTuple(tuple.NewStringValue("persistent"))
	rid, err := db.Insert("users", tup)
	if err != nil {
		t.Fatalf("Insert() error = %v", err)
	}
	db.Close()

	db2, err := database.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db2.Close()

	got, err := db2.GetTuple("users", rid)
	if err != nil {
		t.Fatalf("GetTuple() after reopen error = %v", err)
	}
	if len(got.Values) != 1 {
		t.Fatalf("len(Values) = %d, want 1", len(got.Values))
	}
	v := got.Values[0].(*tuple.StringValue)
	if v.Value != "persistent" {
		t.Errorf("value = %q, want %q", v.Value, "persistent")
	}
}

func TestOpenHeapFile(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	db, err := database.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer db.Close()

	columns := []catalog.Column{{Name: "id", Type: catalog.TypeInt32Type}}
	err = db.CreateTable("users", columns)
	if err != nil {
		t.Fatalf("CreateTable() error = %v", err)
	}

	hf, err := db.OpenHeapFile("users")
	if err != nil {
		t.Fatalf("OpenHeapFile() error = %v, want nil", err)
	}

	if hf == nil {
		t.Fatalf("OpenHeapFile() returned nil")
	}
}

func TestOpenHeapFileTableNotFound(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	db, err := database.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer db.Close()

	_, err = db.OpenHeapFile("nonexistent")
	if !errors.Is(err, database.ErrTableNotFound) {
		t.Errorf("OpenHeapFile() error = %v, want %v", err, database.ErrTableNotFound)
	}
}

func TestOpenHeapFileEmptyName(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	db, err := database.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer db.Close()

	_, err = db.OpenHeapFile("")
	if !errors.Is(err, database.ErrInvalidTableName) {
		t.Errorf("OpenHeapFile() error = %v, want %v", err, database.ErrInvalidTableName)
	}
}

func TestDatabaseInsert(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	db, err := database.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer db.Close()

	columns := []catalog.Column{{Name: "id", Type: catalog.TypeInt32Type}}
	err = db.CreateTable("users", columns)
	if err != nil {
		t.Fatalf("CreateTable() error = %v", err)
	}

	tup := tuple.NewTuple(tuple.NewInt32Value(100))
	rid, err := db.Insert("users", tup)
	if err != nil {
		t.Fatalf("Insert() error = %v, want nil", err)
	}

	if rid == nil {
		t.Fatalf("Insert() returned nil RecordID")
	}

	if rid.PageID != 1 {
		t.Errorf("PageID = %d, want 1", rid.PageID)
	}

	if rid.SlotID != 0 {
		t.Errorf("SlotID = %d, want 0", rid.SlotID)
	}
}

func TestDatabaseInsertAndGetTuple(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	db, err := database.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer db.Close()

	columns := []catalog.Column{{Name: "name", Type: catalog.TypeStringType}}
	err = db.CreateTable("users", columns)
	if err != nil {
		t.Fatalf("CreateTable() error = %v", err)
	}

	tup := tuple.NewTuple(tuple.NewStringValue("Alice"))
	rid, err := db.Insert("users", tup)
	if err != nil {
		t.Fatalf("Insert() error = %v", err)
	}

	got, err := db.GetTuple("users", rid)
	if err != nil {
		t.Fatalf("GetTuple() error = %v, want nil", err)
	}

	if len(got.Values) != 1 {
		t.Fatalf("len(Values) = %d, want 1", len(got.Values))
	}
	v := got.Values[0].(*tuple.StringValue)
	if v.Value != "Alice" {
		t.Errorf("value = %q, want %q", v.Value, "Alice")
	}
}

func TestDatabaseInsertMultipleRecords(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	db, err := database.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer db.Close()

	columns := []catalog.Column{{Name: "name", Type: catalog.TypeStringType}}
	err = db.CreateTable("users", columns)
	if err != nil {
		t.Fatalf("CreateTable() error = %v", err)
	}

	records := []struct {
		name  string
		value string
	}{
		{"Alice", "Alice"},
		{"Bob", "Bob"},
		{"Charlie", "Charlie"},
	}

	var rids []*heapfile.RecordID
	for _, rec := range records {
		tup := tuple.NewTuple(tuple.NewStringValue(rec.value))
		rid, err := db.Insert("users", tup)
		if err != nil {
			t.Fatalf("Insert(%q) error = %v", rec.value, err)
		}
		rids = append(rids, rid)
	}

	for i, rid := range rids {
		got, err := db.GetTuple("users", rid)
		if err != nil {
			t.Fatalf("GetTuple(%d) error = %v", i, err)
		}
		if len(got.Values) != 1 {
			t.Fatalf("len(Values) = %d, want 1", len(got.Values))
		}
		v := got.Values[0].(*tuple.StringValue)
		if v.Value != records[i].value {
			t.Errorf("value[%d] = %q, want %q", i, v.Value, records[i].value)
		}
	}
}

func TestDatabaseInsertPersistsAfterReopen(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	db, err := database.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	columns := []catalog.Column{{Name: "name", Type: catalog.TypeStringType}}
	err = db.CreateTable("users", columns)
	if err != nil {
		t.Fatalf("CreateTable() error = %v", err)
	}

	tup := tuple.NewTuple(tuple.NewStringValue("Persistent data"))
	rid, err := db.Insert("users", tup)
	if err != nil {
		t.Fatalf("Insert() error = %v", err)
	}

	got, err := db.GetTuple("users", rid)
	if err != nil {
		t.Fatalf("GetTuple() error = %v", err)
	}
	if len(got.Values) != 1 {
		t.Fatalf("len(Values) = %d, want 1", len(got.Values))
	}
	v := got.Values[0].(*tuple.StringValue)
	if v.Value != "Persistent data" {
		t.Errorf("value = %q, want %q", v.Value, "Persistent data")
	}

	db.Close()

	db2, err := database.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db2.Close()

	got2, err := db2.GetTuple("users", rid)
	if err != nil {
		t.Fatalf("GetTuple() after reopen error = %v", err)
	}
	if len(got2.Values) != 1 {
		t.Fatalf("len(Values) = %d, want 1", len(got2.Values))
	}
	v2 := got2.Values[0].(*tuple.StringValue)
	if v2.Value != "Persistent data" {
		t.Errorf("value after reopen = %q, want %q", v2.Value, "Persistent data")
	}
}

func TestDatabaseInsertTableNotFound(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	db, err := database.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer db.Close()

	tup := tuple.NewTuple(tuple.NewInt32Value(1))
	_, err = db.Insert("nonexistent", tup)
	if !errors.Is(err, database.ErrTableNotFound) {
		t.Errorf("Insert() error = %v, want %v", err, database.ErrTableNotFound)
	}
}

func TestDatabaseInsertEmptyTableName(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	db, err := database.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer db.Close()

	tup := tuple.NewTuple(tuple.NewInt32Value(1))
	_, err = db.Insert("", tup)
	if !errors.Is(err, database.ErrInvalidTableName) {
		t.Errorf("Insert() error = %v, want %v", err, database.ErrInvalidTableName)
	}
}

func TestDatabaseGetTupleTableNotFound(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	db, err := database.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer db.Close()

	rid := &heapfile.RecordID{PageID: 1, SlotID: 0}
	_, err = db.GetTuple("nonexistent", rid)
	if !errors.Is(err, database.ErrTableNotFound) {
		t.Errorf("GetTuple() error = %v, want %v", err, database.ErrTableNotFound)
	}
}

func TestDatabaseGetTupleEmptyTableName(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	db, err := database.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer db.Close()

	_, err = db.GetTuple("", &heapfile.RecordID{PageID: 1, SlotID: 0})
	if !errors.Is(err, database.ErrInvalidTableName) {
		t.Errorf("GetTuple() error = %v, want %v", err, database.ErrInvalidTableName)
	}
}

func TestDatabaseGetTupleNilRID(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	db, err := database.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer db.Close()

	columns := []catalog.Column{{Name: "name", Type: catalog.TypeStringType}}
	err = db.CreateTable("users", columns)
	if err != nil {
		t.Fatalf("CreateTable() error = %v", err)
	}

	_, err = db.GetTuple("users", nil)
	if !errors.Is(err, database.ErrInvalidRecordID) {
		t.Errorf("GetTuple() error = %v, want %v", err, database.ErrInvalidRecordID)
	}
}

func TestDatabaseInsertWrongValueCount(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	db, err := database.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer db.Close()

	err = db.CreateTable("users", []catalog.Column{
		{Name: "id", Type: catalog.TypeInt32Type},
		{Name: "name", Type: catalog.TypeStringType},
	})
	if err != nil {
		t.Fatalf("CreateTable() error = %v", err)
	}

	// Only 1 value, table expects 2
	tup := tuple.NewTuple(tuple.NewStringValue("Alice"))
	_, err = db.Insert("users", tup)
	if err != tuple.ErrInvalidTuple {
		t.Errorf("Insert() error = %v, want %v", err, tuple.ErrInvalidTuple)
	}
}

func TestDatabaseInsertWrongValueType(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	db, err := database.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer db.Close()

	err = db.CreateTable("users", []catalog.Column{
		{Name: "name", Type: catalog.TypeStringType},
	})
	if err != nil {
		t.Fatalf("CreateTable() error = %v", err)
	}

	// Int32Value where StringValue is expected
	tup := tuple.NewTuple(tuple.NewInt32Value(42))
	_, err = db.Insert("users", tup)
	if err != tuple.ErrInvalidTuple {
		t.Errorf("Insert() error = %v, want %v", err, tuple.ErrInvalidTuple)
	}
}

func TestDatabaseInsertAndGetDifferentTables(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	db, err := database.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer db.Close()

	err = db.CreateTable("users", []catalog.Column{{Name: "name", Type: catalog.TypeStringType}})
	if err != nil {
		t.Fatalf("CreateTable('users') error = %v", err)
	}

	err = db.CreateTable("products", []catalog.Column{{Name: "title", Type: catalog.TypeStringType}})
	if err != nil {
		t.Fatalf("CreateTable('products') error = %v", err)
	}

	tup1 := tuple.NewTuple(tuple.NewStringValue("Alice"))
	rid1, err := db.Insert("users", tup1)
	if err != nil {
		t.Fatalf("Insert('users') error = %v", err)
	}

	tup2 := tuple.NewTuple(tuple.NewStringValue("Widget"))
	rid2, err := db.Insert("products", tup2)
	if err != nil {
		t.Fatalf("Insert('products') error = %v", err)
	}

	got1, err := db.GetTuple("users", rid1)
	if err != nil {
		t.Fatalf("GetTuple('users') error = %v", err)
	}
	if len(got1.Values) != 1 {
		t.Fatalf("len(Values) = %d, want 1", len(got1.Values))
	}
	v1 := got1.Values[0].(*tuple.StringValue)
	if v1.Value != "Alice" {
		t.Errorf("value = %q, want %q", v1.Value, "Alice")
	}

	got2, err := db.GetTuple("products", rid2)
	if err != nil {
		t.Fatalf("GetTuple('products') error = %v", err)
	}
	if len(got2.Values) != 1 {
		t.Fatalf("len(Values) = %d, want 1", len(got2.Values))
	}
	v2 := got2.Values[0].(*tuple.StringValue)
	if v2.Value != "Widget" {
		t.Errorf("value = %q, want %q", v2.Value, "Widget")
	}
}
