package storage_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"

	"db-engine-v1/internal/storage"
	"db-engine-v1/internal/storage/database"
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
