package storage_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"db-engine-v1/internal/storage/filemanager"
)

func TestOpen(t *testing.T) {
	t.Run("creates new file if it does not exist", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "new.db")

		fm, err := filemanager.Open(path)
		if err != nil {
			t.Fatalf("Open() error = %v, want nil", err)
		}
		defer fm.Close()

		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file to exist at %s, got error: %v", path, err)
		}
	})

	t.Run("opens existing file without truncating", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "existing.db")

		if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		fm, err := filemanager.Open(path)
		if err != nil {
			t.Fatalf("Open() error = %v, want nil", err)
		}
		defer fm.Close()

		size, err := fm.Size()
		if err != nil {
			t.Fatalf("Size() error = %v, want nil", err)
		}
		if size != 5 {
			t.Errorf("Size() = %d, want %d (existing content should be preserved)", size, 5)
		}
	})

	t.Run("returns error for invalid path", func(t *testing.T) {
		// A path inside a non-existent directory cannot be created.
		path := filepath.Join(t.TempDir(), "missing-dir", "file.db")

		fm, err := filemanager.Open(path)
		if err == nil {
			fm.Close()
			t.Fatal("Open() error = nil, want non-nil")
		}
		if fm != nil {
			t.Errorf("Open() returned non-nil FileManager on error: %v", fm)
		}
	})
}

func TestFileManager_WriteAt_ReadAt(t *testing.T) {
	fm := openTempFileManager(t)

	want := []byte("hello, database!")

	if err := fm.WriteAt(0, want); err != nil {
		t.Fatalf("WriteAt() error = %v, want nil", err)
	}

	got := make([]byte, len(want))
	if err := fm.ReadAt(0, got); err != nil {
		t.Fatalf("ReadAt() error = %v, want nil", err)
	}

	if !bytes.Equal(got, want) {
		t.Errorf("ReadAt() = %q, want %q", got, want)
	}
}

func TestFileManager_WriteAt_NonZeroOffset(t *testing.T) {
	fm := openTempFileManager(t)

	if err := fm.WriteAt(10, []byte("data")); err != nil {
		t.Fatalf("WriteAt() error = %v, want nil", err)
	}

	got := make([]byte, 4)
	if err := fm.ReadAt(10, got); err != nil {
		t.Fatalf("ReadAt() error = %v, want nil", err)
	}

	if !bytes.Equal(got, []byte("data")) {
		t.Errorf("ReadAt() = %q, want %q", got, "data")
	}

	size, err := fm.Size()
	if err != nil {
		t.Fatalf("Size() error = %v, want nil", err)
	}
	if size != 14 {
		t.Errorf("Size() = %d, want %d", size, 14)
	}
}

func TestFileManager_ReadAt_ShortRead(t *testing.T) {
	fm := openTempFileManager(t)

	// Write fewer bytes than we will attempt to read.
	if err := fm.WriteAt(0, []byte("abc")); err != nil {
		t.Fatalf("WriteAt() error = %v, want nil", err)
	}

	buffer := make([]byte, 10)
	err := fm.ReadAt(0, buffer)

	if !errors.Is(err, filemanager.ErrShortRead) {
		t.Errorf("ReadAt() error = %v, want %v", err, filemanager.ErrShortRead)
	}
}

func TestFileManager_ReadAt_EmptyFile(t *testing.T) {
	fm := openTempFileManager(t)

	buffer := make([]byte, 4)
	err := fm.ReadAt(0, buffer)

	if !errors.Is(err, filemanager.ErrShortRead) {
		t.Errorf("ReadAt() error = %v, want %v", err, filemanager.ErrShortRead)
	}
}

func TestFileManager_ReadAt_ZeroLengthBuffer(t *testing.T) {
	fm := openTempFileManager(t)

	buffer := make([]byte, 0)
	if err := fm.ReadAt(0, buffer); err != nil {
		t.Errorf("ReadAt() error = %v, want nil", err)
	}
}

func TestFileManager_ReadAt_ClosedFile(t *testing.T) {
	fm := openTempFileManager(t)
	fm.Close()

	buffer := make([]byte, 4)
	err := fm.ReadAt(0, buffer)

	if !errors.Is(err, filemanager.ErrShortRead) {
		t.Errorf("ReadAt() error = %v, want %v", err, filemanager.ErrShortRead)
	}
}

func TestFileManager_WriteAt_ClosedFile(t *testing.T) {
	fm := openTempFileManager(t)
	fm.Close()

	err := fm.WriteAt(0, []byte("data"))

	if !errors.Is(err, filemanager.ErrShortWrite) {
		t.Errorf("WriteAt() error = %v, want %v", err, filemanager.ErrShortWrite)
	}
}

func TestFileManager_WriteAt_ZeroLengthBuffer(t *testing.T) {
	fm := openTempFileManager(t)

	buffer := make([]byte, 0)
	if err := fm.WriteAt(0, buffer); err != nil {
		t.Errorf("WriteAt() error = %v, want nil", err)
	}
}

func TestFileManager_Sync(t *testing.T) {
	t.Run("succeeds on open file", func(t *testing.T) {
		fm := openTempFileManager(t)

		if err := fm.WriteAt(0, []byte("data")); err != nil {
			t.Fatalf("WriteAt() error = %v, want nil", err)
		}

		if err := fm.Sync(); err != nil {
			t.Errorf("Sync() error = %v, want nil", err)
		}
	})

	t.Run("returns error on closed file", func(t *testing.T) {
		fm := openTempFileManager(t)
		fm.Close()

		if err := fm.Sync(); err == nil {
			t.Error("Sync() error = nil, want non-nil")
		}
	})
}

func TestFileManager_Size(t *testing.T) {
	t.Run("empty file", func(t *testing.T) {
		fm := openTempFileManager(t)

		size, err := fm.Size()
		if err != nil {
			t.Fatalf("Size() error = %v, want nil", err)
		}
		if size != 0 {
			t.Errorf("Size() = %d, want %d", size, 0)
		}
	})

	t.Run("after writing data", func(t *testing.T) {
		fm := openTempFileManager(t)

		data := []byte("0123456789")
		if err := fm.WriteAt(0, data); err != nil {
			t.Fatalf("WriteAt() error = %v, want nil", err)
		}

		size, err := fm.Size()
		if err != nil {
			t.Fatalf("Size() error = %v, want nil", err)
		}
		if size != int64(len(data)) {
			t.Errorf("Size() = %d, want %d", size, len(data))
		}
	})

	t.Run("returns error on closed file", func(t *testing.T) {
		fm := openTempFileManager(t)
		fm.Close()

		if _, err := fm.Size(); err == nil {
			t.Error("Size() error = nil, want non-nil")
		}
	})
}

// openTempFileManager opens a FileManager backed by a fresh file in a
// temporary directory that is cleaned up automatically when the test ends.
func openTempFileManager(t *testing.T) *filemanager.FileManager {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	fm, err := filemanager.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v, want nil", err)
	}

	t.Cleanup(func() {
		fm.Close()
	})

	return fm
}
