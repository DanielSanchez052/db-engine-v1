package storage_test

import (
	"bytes"
	"testing"

	"db-engine-v1/internal/storage/filemanager"
	"db-engine-v1/internal/storage/page"
	"db-engine-v1/internal/storage/pager"
)

func TestNewPager(t *testing.T) {
	fm := &filemanager.FileManager{}
	p := pager.NewPager(fm)

	if p == nil {
		t.Fatalf("NewPager() returned nil, want non-nil")
	}
}

func TestSaveAndLoadPage(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	fm, err := filemanager.Open(path)
	if err != nil {
		t.Fatalf("filemanager.Open() error = %v, want nil", err)
	}
	defer fm.Close()

	p := pager.NewPager(fm)

	// Create a test page
	testPage := page.NewPage(1, page.DataPage)
	testPage.Payload[0] = 42
	testPage.Payload[100] = 99

	// Save the page
	err = p.SavePage(testPage)
	if err != nil {
		t.Fatalf("SavePage() error = %v, want nil", err)
	}

	// Load the page
	loadedPage, err := p.LoadPage(1)
	if err != nil {
		t.Fatalf("LoadPage() error = %v, want nil", err)
	}

	// Verify the page was loaded correctly
	if *loadedPage.Header != *testPage.Header {
		t.Errorf("Header mismatch: got %+v, want %+v", loadedPage.Header, testPage.Header)
	}

	if !bytes.Equal(loadedPage.Payload[:], testPage.Payload[:]) {
		t.Errorf("Payload mismatch: got %v, want %v", loadedPage.Payload, testPage.Payload)
	}
}

func TestLoadPageNotFound(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	fm, err := filemanager.Open(path)
	if err != nil {
		t.Fatalf("filemanager.Open() error = %v, want nil", err)
	}
	defer fm.Close()

	p := pager.NewPager(fm)

	// Try to load a page that hasn't been saved yet
	_, err = p.LoadPage(1)
	if err == nil {
		t.Errorf("LoadPage() expected error for unsaved page, got nil")
	}
}

func TestSavePageError(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	fm, err := filemanager.Open(path)
	if err != nil {
		t.Fatalf("filemanager.Open() error = %v, want nil", err)
	}
	defer fm.Close()

	p := pager.NewPager(fm)

	// Create a test page
	testPage := page.NewPage(1, page.DataPage)

	// Close the file manager to simulate an error
	fm.Close()

	// Try to save the page
	err = p.SavePage(testPage)
	if err == nil {
		t.Errorf("SavePage() expected error for closed file, got nil")
	}
}

func TestMultiplePages(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	fm, err := filemanager.Open(path)
	if err != nil {
		t.Fatalf("filemanager.Open() error = %v, want nil", err)
	}
	defer fm.Close()

	p := pager.NewPager(fm)

	// Create and save multiple pages
	for i := uint64(1); i <= 5; i++ {
		testPage := page.NewPage(i, page.DataPage)
		testPage.Payload[0] = byte(i)

		err = p.SavePage(testPage)
		if err != nil {
			t.Fatalf("SavePage(%d) error = %v, want nil", i, err)
		}
	}

	// Load all pages and verify
	for i := uint64(1); i <= 5; i++ {
		loadedPage, err := p.LoadPage(i)
		if err != nil {
			t.Fatalf("LoadPage(%d) error = %v, want nil", i, err)
		}

		if loadedPage.Header.PageID != i {
			t.Errorf("PageID mismatch: got %d, want %d", loadedPage.Header.PageID, i)
		}

		if loadedPage.Payload[0] != byte(i) {
			t.Errorf("Payload[0] mismatch for page %d: got %d, want %d", i, loadedPage.Payload[0], i)
		}
	}
}
