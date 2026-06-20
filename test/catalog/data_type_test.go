package catalog_test

import (
	"testing"

	"db-engine-v1/internal/catalog"
)

func TestDataTypeValues(t *testing.T) {
	if catalog.TypeInt != 0 {
		t.Errorf("TypeInt = %d, want 0", catalog.TypeInt)
	}

	if catalog.TypeString != 1 {
		t.Errorf("TypeString = %d, want 1", catalog.TypeString)
	}

	if catalog.TypeBool != 2 {
		t.Errorf("TypeBool = %d, want 2", catalog.TypeBool)
	}
}

func TestDataTypeIsValid(t *testing.T) {
	t.Run("TypeInt", func(t *testing.T) {
		if !catalog.TypeInt.IsValid() {
			t.Errorf("TypeInt.IsValid() = false, want true")
		}
	})

	t.Run("TypeString", func(t *testing.T) {
		if !catalog.TypeString.IsValid() {
			t.Errorf("TypeString.IsValid() = false, want true")
		}
	})

	t.Run("TypeBool", func(t *testing.T) {
		if !catalog.TypeBool.IsValid() {
			t.Errorf("TypeBool.IsValid() = false, want true")
		}
	})
}

func TestDataTypeIsValidOutOfRange(t *testing.T) {
	t.Run("below range", func(t *testing.T) {
		var dt catalog.DataType = 255
		if dt.IsValid() {
			t.Errorf("DataType(255).IsValid() = true, want false")
		}
	})

	t.Run("above range", func(t *testing.T) {
		var dt catalog.DataType = 3
		if dt.IsValid() {
			t.Errorf("DataType(3).IsValid() = true, want false")
		}
	})
}
