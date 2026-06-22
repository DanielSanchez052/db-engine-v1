package catalog_test

import (
	"testing"

	"db-engine-v1/internal/catalog"
)

func TestTableNew(t *testing.T) {
	table := catalog.Table{
		Name:     "users",
		HeapName: "users_heap",
		Columns: []catalog.Column{
			{Name: "id", Type: catalog.TypeInt32Type},
			{Name: "name", Type: catalog.TypeStringType},
		},
	}

	if table.Name != "users" {
		t.Errorf("Name = %q, want %q", table.Name, "users")
	}

	if table.HeapName != "users_heap" {
		t.Errorf("HeapName = %q, want %q", table.HeapName, "users_heap")
	}

	if len(table.Columns) != 2 {
		t.Fatalf("len(Columns) = %d, want 2", len(table.Columns))
	}
}

func TestTableSize(t *testing.T) {
	t.Run("no columns", func(t *testing.T) {
		table := catalog.Table{Name: "t", HeapName: "h"}
		expected := 2 + 1 + 2 + 1 + 2
		if table.Size() != expected {
			t.Errorf("Size() = %d, want %d", table.Size(), expected)
		}
	})

	t.Run("with columns", func(t *testing.T) {
		table := catalog.Table{
			Name:     "users",
			HeapName: "users_heap",
			Columns: []catalog.Column{
				{Name: "id", Type: catalog.TypeInt32Type},
				{Name: "name", Type: catalog.TypeStringType},
			},
		}
		// name "users" = 2+5, heap "users_heap" = 2+10, colCount = 2
		// col "id" = 2+2+1 = 5, preceded by ColumnLengthSize = 2
		// col "name" = 2+4+1 = 7, preceded by ColumnLengthSize = 2
		expected := 2 + 5 + 2 + 10 + 2 + (2 + 5) + (2 + 7)
		if table.Size() != expected {
			t.Errorf("Size() = %d, want %d", table.Size(), expected)
		}
	})
}

func TestTableSerialize(t *testing.T) {
	table := catalog.Table{
		Name:     "users",
		HeapName: "users_heap",
		Columns: []catalog.Column{
			{Name: "id", Type: catalog.TypeInt32Type},
		},
	}

	data, err := table.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v, want nil", err)
	}

	if len(data) != table.Size() {
		t.Errorf("len(data) = %d, want %d", len(data), table.Size())
	}
}

func TestNewTableFromBytes(t *testing.T) {
	original := catalog.Table{
		Name:     "products",
		HeapName: "products_heap",
		Columns: []catalog.Column{
			{Name: "id", Type: catalog.TypeInt32Type},
			{Name: "name", Type: catalog.TypeStringType},
			{Name: "price", Type: catalog.TypeInt32Type},
		},
	}

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	got, err := catalog.NewTableFromBytes(data)
	if err != nil {
		t.Fatalf("NewTableFromBytes() error = %v", err)
	}

	if got.Name != original.Name {
		t.Errorf("Name = %q, want %q", got.Name, original.Name)
	}

	if got.HeapName != original.HeapName {
		t.Errorf("HeapName = %q, want %q", got.HeapName, original.HeapName)
	}

	if len(got.Columns) != len(original.Columns) {
		t.Fatalf("len(Columns) = %d, want %d", len(got.Columns), len(original.Columns))
	}

	for i, col := range original.Columns {
		if got.Columns[i].Name != col.Name {
			t.Errorf("Columns[%d].Name = %q, want %q", i, got.Columns[i].Name, col.Name)
		}
		if got.Columns[i].Type != col.Type {
			t.Errorf("Columns[%d].Type = %v, want %v", i, got.Columns[i].Type, col.Type)
		}
	}
}

func TestTableRoundTrip(t *testing.T) {
	original := catalog.Table{
		Name:     "orders",
		HeapName: "orders_heap",
		Columns: []catalog.Column{
			{Name: "id", Type: catalog.TypeInt32Type},
			{Name: "total", Type: catalog.TypeInt32Type},
			{Name: "description", Type: catalog.TypeStringType},
			{Name: "is_shipped", Type: catalog.TypeBoolType},
		},
	}

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	got, err := catalog.NewTableFromBytes(data)
	if err != nil {
		t.Fatalf("NewTableFromBytes() error = %v", err)
	}

	if got.Name != original.Name {
		t.Errorf("Name = %q, want %q", got.Name, original.Name)
	}

	if got.HeapName != original.HeapName {
		t.Errorf("HeapName = %q, want %q", got.HeapName, original.HeapName)
	}

	if len(got.Columns) != len(original.Columns) {
		t.Fatalf("len(Columns) = %d, want %d", len(got.Columns), len(original.Columns))
	}

	for i := range original.Columns {
		if got.Columns[i].Name != original.Columns[i].Name {
			t.Errorf("Columns[%d].Name = %q, want %q", i, got.Columns[i].Name, original.Columns[i].Name)
		}
		if got.Columns[i].Type != original.Columns[i].Type {
			t.Errorf("Columns[%d].Type = %v, want %v", i, got.Columns[i].Type, original.Columns[i].Type)
		}
	}
}

func TestTableSerializeEmptyNames(t *testing.T) {
	table := catalog.Table{
		Name:     "",
		HeapName: "",
		Columns: []catalog.Column{
			{Name: "col", Type: catalog.TypeInt32Type},
		},
	}

	data, err := table.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	got, err := catalog.NewTableFromBytes(data)
	if err != nil {
		t.Fatalf("NewTableFromBytes() error = %v", err)
	}

	if got.Name != "" {
		t.Errorf("Name = %q, want empty", got.Name)
	}

	if got.HeapName != "" {
		t.Errorf("HeapName = %q, want empty", got.HeapName)
	}
}

func TestTableSerializeNoColumns(t *testing.T) {
	table := catalog.Table{
		Name:     "empty_table",
		HeapName: "empty_heap",
		Columns:  []catalog.Column{},
	}

	data, err := table.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	got, err := catalog.NewTableFromBytes(data)
	if err != nil {
		t.Fatalf("NewTableFromBytes() error = %v", err)
	}

	if got.Name != "empty_table" {
		t.Errorf("Name = %q, want %q", got.Name, "empty_table")
	}

	if len(got.Columns) != 0 {
		t.Errorf("len(Columns) = %d, want 0", len(got.Columns))
	}
}

func TestTableSerializeSingleColumn(t *testing.T) {
	table := catalog.Table{
		Name:     "single",
		HeapName: "single_heap",
		Columns: []catalog.Column{
			{Name: "id", Type: catalog.TypeInt32Type},
		},
	}

	data, err := table.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	got, err := catalog.NewTableFromBytes(data)
	if err != nil {
		t.Fatalf("NewTableFromBytes() error = %v", err)
	}

	if len(got.Columns) != 1 {
		t.Fatalf("len(Columns) = %d, want 1", len(got.Columns))
	}

	if got.Columns[0].Name != "id" || got.Columns[0].Type != catalog.TypeInt32Type {
		t.Errorf("Column = %+v, want {Name:id Type:TypeInt}", got.Columns[0])
	}
}

func TestTableSerializeAllTypes(t *testing.T) {
	table := catalog.Table{
		Name:     "all_types",
		HeapName: "all_types_heap",
		Columns: []catalog.Column{
			{Name: "a", Type: catalog.TypeInt32Type},
			{Name: "b", Type: catalog.TypeStringType},
			{Name: "c", Type: catalog.TypeBoolType},
		},
	}

	data, err := table.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	got, err := catalog.NewTableFromBytes(data)
	if err != nil {
		t.Fatalf("NewTableFromBytes() error = %v", err)
	}

	types := []catalog.DataType{catalog.TypeInt32Type, catalog.TypeStringType, catalog.TypeBoolType}
	for i, dt := range types {
		if got.Columns[i].Type != dt {
			t.Errorf("Columns[%d].Type = %v, want %v", i, got.Columns[i].Type, dt)
		}
	}
}

func TestTableSerializeLongNames(t *testing.T) {
	longName := ""
	for i := 0; i < 255; i++ {
		longName += "a"
	}

	table := catalog.Table{
		Name:     longName,
		HeapName: longName,
		Columns: []catalog.Column{
			{Name: "id", Type: catalog.TypeInt32Type},
		},
	}

	data, err := table.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	got, err := catalog.NewTableFromBytes(data)
	if err != nil {
		t.Fatalf("NewTableFromBytes() error = %v", err)
	}

	if len(got.Name) != 255 {
		t.Errorf("Name length = %d, want 255", len(got.Name))
	}

	if len(got.HeapName) != 255 {
		t.Errorf("HeapName length = %d, want 255", len(got.HeapName))
	}
}

func TestNewTableFromBytesInvalid(t *testing.T) {
	t.Run("empty data", func(t *testing.T) {
		_, err := catalog.NewTableFromBytes([]byte{})
		if err == nil {
			t.Errorf("NewTableFromBytes() expected error for empty data, got nil")
		}
	})

	t.Run("truncated name", func(t *testing.T) {
		data := make([]byte, 1)
		_, err := catalog.NewTableFromBytes(data)
		if err == nil {
			t.Errorf("NewTableFromBytes() expected error for truncated name, got nil")
		}
	})

	t.Run("truncated heap name", func(t *testing.T) {
		data := []byte{
			0x01, 0x00, // name length = 1
			'a',        // name = "a"
			0x05, 0x00, // heap name length = 5
			'b', 'c', // only 2 bytes but heap name expects 5
		}
		_, err := catalog.NewTableFromBytes(data)
		if err == nil {
			t.Errorf("NewTableFromBytes() expected error for truncated heap name, got nil")
		}
	})

	t.Run("truncated column count", func(t *testing.T) {
		data := []byte{
			0x01, 0x00, // name length = 1
			'a',        // name = "a"
			0x01, 0x00, // heap name length = 1
			'b', // heap name = "b"
			// missing column count
		}
		_, err := catalog.NewTableFromBytes(data)
		if err == nil {
			t.Errorf("NewTableFromBytes() expected error for truncated column count, got nil")
		}
	})

	t.Run("truncated column data", func(t *testing.T) {
		data := []byte{
			0x01, 0x00, // name length = 1
			'a',        // name = "a"
			0x01, 0x00, // heap name length = 1
			'b',        // heap name = "b"
			0x01, 0x00, // column count = 1
			0x05, 0x00, // column length = 5
			0x01, 0x00, // column name length = 1
			'c', // column name = "c"
			// missing column type byte (should be 1 more byte)
		}
		_, err := catalog.NewTableFromBytes(data)
		if err == nil {
			t.Errorf("NewTableFromBytes() expected error for truncated column data, got nil")
		}
	})
}
