package catalog_test

import (
	"testing"

	"db-engine-v1/internal/catalog"
)

func TestColumnNew(t *testing.T) {
	c := catalog.Column{
		Name: "age",
		Type: catalog.TypeInt32Type,
	}

	if c.Name != "age" {
		t.Errorf("Name = %q, want %q", c.Name, "age")
	}

	if c.Type != catalog.TypeInt32Type {
		t.Errorf("Type = %v, want %v", c.Type, catalog.TypeInt32Type)
	}
}

func TestColumnSize(t *testing.T) {
	t.Run("short name", func(t *testing.T) {
		c := catalog.Column{Name: "id", Type: catalog.TypeInt32Type}
		expected := 2 + 2 + 1
		if c.Size() != expected {
			t.Errorf("Size() = %d, want %d", c.Size(), expected)
		}
	})

	t.Run("long name", func(t *testing.T) {
		c := catalog.Column{Name: "full_name", Type: catalog.TypeStringType}
		expected := 2 + 9 + 1
		if c.Size() != expected {
			t.Errorf("Size() = %d, want %d", c.Size(), expected)
		}
	})
}

func TestColumnSerialize(t *testing.T) {
	c := catalog.Column{Name: "age", Type: catalog.TypeInt32Type}

	data, err := c.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v, want nil", err)
	}

	if len(data) != c.Size() {
		t.Errorf("len(data) = %d, want %d", len(data), c.Size())
	}
}

func TestNewColumnFromBytes(t *testing.T) {
	original := catalog.Column{Name: "name", Type: catalog.TypeStringType}

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v, want nil", err)
	}

	got, err := catalog.NewColumnFromBytes(data)
	if err != nil {
		t.Fatalf("NewColumnFromBytes() error = %v, want nil", err)
	}

	if got.Name != original.Name {
		t.Errorf("Name = %q, want %q", got.Name, original.Name)
	}

	if got.Type != original.Type {
		t.Errorf("Type = %v, want %v", got.Type, original.Type)
	}
}

func TestColumnRoundTrip(t *testing.T) {
	t.Run("TypeInt32Type", func(t *testing.T) {
		original := catalog.Column{Name: "id", Type: catalog.TypeInt32Type}
		data, _ := original.Serialize()
		got, err := catalog.NewColumnFromBytes(data)
		if err != nil {
			t.Fatalf("NewColumnFromBytes() error = %v", err)
		}
		if got.Name != "id" || got.Type != catalog.TypeInt32Type {
			t.Errorf("Round trip failed: got %+v", got)
		}
	})

	t.Run("TypeStringType", func(t *testing.T) {
		original := catalog.Column{Name: "username", Type: catalog.TypeStringType}
		data, _ := original.Serialize()
		got, err := catalog.NewColumnFromBytes(data)
		if err != nil {
			t.Fatalf("NewColumnFromBytes() error = %v", err)
		}
		if got.Name != "username" || got.Type != catalog.TypeStringType {
			t.Errorf("Round trip failed: got %+v", got)
		}
	})

	t.Run("TypeBoolType", func(t *testing.T) {
		original := catalog.Column{Name: "is_active", Type: catalog.TypeBoolType}
		data, _ := original.Serialize()
		got, err := catalog.NewColumnFromBytes(data)
		if err != nil {
			t.Fatalf("NewColumnFromBytes() error = %v", err)
		}
		if got.Name != "is_active" || got.Type != catalog.TypeBoolType {
			t.Errorf("Round trip failed: got %+v", got)
		}
	})
}

func TestColumnSerializeInvalidType(t *testing.T) {
	c := catalog.Column{Name: "bad", Type: catalog.DataType(99)}

	_, err := c.Serialize()
	if err == nil {
		t.Errorf("Serialize() expected error for invalid type, got nil")
	}
}

func TestColumnSerializeEmptyName(t *testing.T) {
	c := catalog.Column{Name: "", Type: catalog.TypeStringType}

	data, err := c.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v, want nil", err)
	}

	got, err := catalog.NewColumnFromBytes(data)
	if err != nil {
		t.Fatalf("NewColumnFromBytes() error = %v, want nil", err)
	}

	if got.Name != "" {
		t.Errorf("Name = %q, want empty", got.Name)
	}
}

func TestColumnSerializeNameMaxLength(t *testing.T) {
	longName := ""
	for i := 0; i < 255; i++ {
		longName += "a"
	}

	c := catalog.Column{Name: longName, Type: catalog.TypeInt32Type}

	data, err := c.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v, want nil", err)
	}

	got, err := catalog.NewColumnFromBytes(data)
	if err != nil {
		t.Fatalf("NewColumnFromBytes() error = %v, want nil", err)
	}

	if got.Name != longName {
		t.Errorf("Name length mismatch: got %d, want %d", len(got.Name), len(longName))
	}
}

func TestNewColumnFromBytesInvalid(t *testing.T) {
	t.Run("empty data", func(t *testing.T) {
		_, err := catalog.NewColumnFromBytes([]byte{})
		if err == nil {
			t.Errorf("NewColumnFromBytes() expected error for empty data, got nil")
		}
	})

	t.Run("missing type byte", func(t *testing.T) {
		data := make([]byte, 2)
		_, err := catalog.NewColumnFromBytes(data)
		if err == nil {
			t.Errorf("NewColumnFromBytes() expected error for truncated data, got nil")
		}
	})

	t.Run("invalid data type", func(t *testing.T) {
		data := []byte{
			0x02, 0x00, // string length = 2
			'a', 'b', // name = "ab"
			0x05, // type = 5 (invalid)
		}
		_, err := catalog.NewColumnFromBytes(data)
		if err == nil {
			t.Errorf("NewColumnFromBytes() expected error for invalid type, got nil")
		}
	})
}
