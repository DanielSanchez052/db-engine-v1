package storage_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"

	"db-engine-v1/internal/storage/page"
)

func TestNewPageHeader(t *testing.T) {
	h := page.NewPageHeader(1, page.DataPage)

	if h.PageID != 1 {
		t.Errorf("PageID = %d, want %d", h.PageID, 1)
	}

	if h.PageType != page.DataPage {
		t.Errorf("PageType = %d, want %d", h.PageType, page.DataPage)
	}

	if h.RecordCount != 0 {
		t.Errorf("RecordCount = %d, want %d", h.RecordCount, 0)
	}

	if h.FreeSpaceOffset != page.PageSize {
		t.Errorf("FreeSpaceOffset = %d, want %d", h.FreeSpaceOffset, page.PageSize)
	}

	if h.SlotCount != 0 {
		t.Errorf("SlotCount = %d, want %d", h.SlotCount, 0)
	}

	wantReserved := [49]byte{}
	if h.Reserved != wantReserved {
		t.Errorf("Reserved = %v, want %v", h.Reserved, wantReserved)
	}
}

func TestPageHeader_Serialize(t *testing.T) {
	h := page.NewPageHeader(42, page.IndexPage)
	h.RecordCount = 10
	h.FreeSpaceOffset = 100
	h.SlotCount = 5
	copy(h.Reserved[:], []byte("test-data"))

	buf, err := h.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v, want nil", err)
	}

	expectedSize := page.PageHeaderSize
	if len(buf) != expectedSize {
		t.Fatalf("Serialize() len = %d, want %d", len(buf), expectedSize)
	}

	if got := binary.LittleEndian.Uint64(buf[page.PageIDOffset : page.PageIDOffset+page.PageIDSize]); got != h.PageID {
		t.Errorf("PageID bytes = %d, want %d", got, h.PageID)
	}

	if got := binary.LittleEndian.Uint16(buf[page.RecordCountOffset : page.RecordCountOffset+page.RecordCountSize]); got != h.RecordCount {
		t.Errorf("RecordCount bytes = %d, want %d", got, h.RecordCount)
	}

	if got := binary.LittleEndian.Uint16(buf[page.FreeSpaceOffsetOffset : page.FreeSpaceOffsetOffset+page.FreeSpaceOffsetSize]); got != h.FreeSpaceOffset {
		t.Errorf("FreeSpaceOffset bytes = %d, want %d", got, h.FreeSpaceOffset)
	}

	if got := binary.LittleEndian.Uint16(buf[page.SlotCountOffset : page.SlotCountOffset+page.SlotCountSize]); got != h.SlotCount {
		t.Errorf("SlotCount bytes = %d, want %d", got, h.SlotCount)
	}

	if got := page.PageType(buf[page.PageTypeOffset]); got != h.PageType {
		t.Errorf("PageType bytes = %d, want %d", got, h.PageType)
	}

	if !bytes.Equal(buf[page.ReservedOffset:page.ReservedOffset+page.ReservedSize], h.Reserved[:]) {
		t.Errorf("Reserved bytes = %v, want %v", buf[page.ReservedOffset:page.ReservedOffset+page.ReservedSize], h.Reserved[:])
	}
}

func TestNewPageHeaderFromBytes(t *testing.T) {
	t.Run("size mismatch", func(t *testing.T) {
		cases := []struct {
			name string
			size int
		}{
			{"too short", page.PageHeaderSize - 1},
			{"too long", page.PageHeaderSize + 1},
			{"empty", 0},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				data := make([]byte, tc.size)
				h, err := page.NewPageHeaderFromBytes(data)
				if h != nil {
					t.Errorf("NewPageHeaderFromBytes() header = %v, want nil", h)
				}
				if !errors.Is(err, page.ErrPageHeaderSizeMismatch) {
					t.Errorf("NewPageHeaderFromBytes() error = %v, want %v", err, page.ErrPageHeaderSizeMismatch)
				}
			})
		}
	})

	t.Run("valid data", func(t *testing.T) {
		original := page.NewPageHeader(123, page.CatalogPage)
		original.RecordCount = 25
		original.FreeSpaceOffset = 500
		original.SlotCount = 12
		copy(original.Reserved[:], []byte("reserved-test"))

		buf, err := original.Serialize()
		if err != nil {
			t.Fatalf("Serialize() error = %v, want nil", err)
		}

		got, err := page.NewPageHeaderFromBytes(buf)
		if err != nil {
			t.Fatalf("NewPageHeaderFromBytes() error = %v, want nil", err)
		}

		if got.PageID != original.PageID {
			t.Errorf("PageID = %d, want %d", got.PageID, original.PageID)
		}
		if got.RecordCount != original.RecordCount {
			t.Errorf("RecordCount = %d, want %d", got.RecordCount, original.RecordCount)
		}
		if got.FreeSpaceOffset != original.FreeSpaceOffset {
			t.Errorf("FreeSpaceOffset = %d, want %d", got.FreeSpaceOffset, original.FreeSpaceOffset)
		}
		if got.SlotCount != original.SlotCount {
			t.Errorf("SlotCount = %d, want %d", got.SlotCount, original.SlotCount)
		}
		if got.PageType != original.PageType {
			t.Errorf("PageType = %d, want %d", got.PageType, original.PageType)
		}
		if got.Reserved != original.Reserved {
			t.Errorf("Reserved = %v, want %v", got.Reserved, original.Reserved)
		}
	})
}

func TestPageHeader_RoundTrip(t *testing.T) {
	original := page.NewPageHeader(999, page.DataPage)

	buf, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v, want nil", err)
	}

	got, err := page.NewPageHeaderFromBytes(buf)
	if err != nil {
		t.Fatalf("NewPageHeaderFromBytes() error = %v, want nil", err)
	}

	if *got != *original {
		t.Errorf("round trip mismatch: got %+v, want %+v", got, original)
	}
}

func TestPageHeader_Validate(t *testing.T) {
	t.Run("valid header", func(t *testing.T) {
		h := page.NewPageHeader(1, page.DataPage)
		if err := h.Validate(); err != nil {
			t.Errorf("Validate() error = %v, want nil", err)
		}
	})

	t.Run("invalid page type", func(t *testing.T) {
		h := page.NewPageHeader(1, page.DataPage)
		h.PageType = 99

		err := h.Validate()
		if !errors.Is(err, page.ErrInvalidPageType) {
			t.Errorf("Validate() error = %v, want %v", err, page.ErrInvalidPageType)
		}
	})

	t.Run("free space offset exceeds page size", func(t *testing.T) {
		h := page.NewPageHeader(1, page.DataPage)
		h.FreeSpaceOffset = page.PageSize + 1

		err := h.Validate()
		if !errors.Is(err, page.ErrInvalidFreeSpaceOffset) {
			t.Errorf("Validate() error = %v, want %v", err, page.ErrInvalidFreeSpaceOffset)
		}
	})

	t.Run("free space offset in header", func(t *testing.T) {
		h := page.NewPageHeader(1, page.DataPage)
		h.FreeSpaceOffset = page.PageHeaderSize - 1

		err := h.Validate()
		if !errors.Is(err, page.ErrFreeSpaceOffsetInHeader) {
			t.Errorf("Validate() error = %v, want %v", err, page.ErrFreeSpaceOffsetInHeader)
		}
	})

	t.Run("invalid page type checked before offset", func(t *testing.T) {
		h := page.NewPageHeader(1, page.DataPage)
		h.PageType = 99
		h.FreeSpaceOffset = page.PageSize + 1

		err := h.Validate()
		if !errors.Is(err, page.ErrInvalidPageType) {
			t.Errorf("Validate() error = %v, want %v", err, page.ErrInvalidPageType)
		}
	})
}

func TestPageType_IsValid(t *testing.T) {
	t.Run("valid page types", func(t *testing.T) {
		validTypes := []page.PageType{page.DataPage, page.IndexPage, page.CatalogPage}
		for _, pt := range validTypes {
			if !pt.IsValid() {
				t.Errorf("PageType %d should be valid", pt)
			}
		}
	})

	t.Run("invalid page type", func(t *testing.T) {
		invalidType := page.PageType(99)
		if invalidType.IsValid() {
			t.Errorf("PageType 99 should be invalid")
		}
	})
}

func TestPageSerialize(t *testing.T) {
	p := page.NewPage(1, page.DataPage)

	bytes, err := p.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v, want nil", err)
	}

	if len(bytes) != page.PageSize {
		t.Errorf("Serialize() len = %d, want %d", len(bytes), page.PageSize)
	}
}

func TestPageDeserialize(t *testing.T) {
	p := page.NewPage(1, page.DataPage)

	bytes, err := p.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v, want nil", err)
	}

	p2, err := page.NewPageFromBytes(bytes)
	if err != nil {
		t.Fatalf("NewPageFromBytes() error = %v, want nil", err)
	}

	if *p.Header != *p2.Header {
		t.Errorf("Header mismatch: got %+v, want %+v", p2.Header, p.Header)
	}
}

func TestPagePayloadPersistence(t *testing.T) {
	p := page.NewPage(1, page.DataPage)

	p.Payload[0] = 10
	p.Payload[100] = 50

	bytes, err := p.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v, want nil", err)
	}

	p2, err := page.NewPageFromBytes(bytes)
	if err != nil {
		t.Fatalf("NewPageFromBytes() error = %v, want nil", err)
	}

	if p2.Payload[0] != 10 {
		t.Errorf("Payload[0] = %d, want 10", p2.Payload[0])
	}

	if p2.Payload[100] != 50 {
		t.Errorf("Payload[100] = %d, want 50", p2.Payload[100])
	}
}

func TestInvalidPageSize(t *testing.T) {
	data := make([]byte, 100)

	_, err := page.NewPageFromBytes(data)
	if err == nil {
		t.Errorf("NewPageFromBytes() expected error for invalid size, got nil")
	}

	if !errors.Is(err, page.ErrPageSizeMismatch) {
		t.Errorf("NewPageFromBytes() error = %v, want %v", err, page.ErrPageSizeMismatch)
	}
}
