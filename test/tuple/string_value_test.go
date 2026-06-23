package tuple_test

import (
	"math"
	"strings"
	"testing"

	"db-engine-v1/internal/catalog"
	"db-engine-v1/internal/storage/tuple"
)

func TestNewStringValue(t *testing.T) {
	v := tuple.NewStringValue("hello")
	if v.Value != "hello" {
		t.Errorf("Value = %q, want %q", v.Value, "hello")
	}
}

func TestStringValueType(t *testing.T) {
	v := tuple.NewStringValue("test")
	if v.Type() != catalog.TypeStringType {
		t.Errorf("Type() = %v, want %v", v.Type(), catalog.TypeStringType)
	}
}

func TestStringValueSize(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  uint16
	}{
		{"empty", "", catalog.StringLengthSize},
		{"ascii", "hello", catalog.StringLengthSize + 5},
		{"tildes", "café", catalog.StringLengthSize + 5},
		{"japanese", "日本", catalog.StringLengthSize + 6},
		{"emoji", "😊", catalog.StringLengthSize + 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tuple.NewStringValue(tt.value)
			got := v.Size()
			if got != tt.want {
				t.Errorf("Size() = %d, want %d (len(value)=%d)", got, tt.want, len(tt.value))
			}
		})
	}
}

func TestStringValueSerialize(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"empty", ""},
		{"ascii", "hello"},
		{"tildes", "café"},
		{"japanese", "日本"},
		{"emoji", "😊"},
		{"mixed", "áéíóúñü ñ"},
		{"emojis_mixed", "hola 😊 mundo 🌍"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tuple.NewStringValue(tt.value)

			data, err := v.Serialize()
			if err != nil {
				t.Fatalf("Serialize() error = %v, want nil", err)
			}

			if len(data) != int(v.Size()) {
				t.Fatalf("len(data) = %d, want %d", len(data), v.Size())
			}

			gotLength := int(data[0]) | int(data[1])<<8
			if gotLength != len(tt.value) {
				t.Errorf("length prefix = %d, want %d", gotLength, len(tt.value))
			}

			gotStr := string(data[catalog.StringLengthSize:])
			if gotStr != tt.value {
				t.Errorf("string content = %q, want %q", gotStr, tt.value)
			}
		})
	}
}

func TestStringValueSerializeTooLong(t *testing.T) {
	v := tuple.NewStringValue(string(make([]byte, math.MaxUint16+1)))
	_, err := v.Serialize()
	if err != tuple.ErrInvalidValue {
		t.Errorf("Serialize() error = %v, want %v", err, tuple.ErrInvalidValue)
	}
}

func TestStringValueFromBytes(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"empty", ""},
		{"ascii", "hello"},
		{"tildes", "café"},
		{"japanese", "日本"},
		{"emoji", "😊"},
		{"mixed", "áéíóúñü ñ"},
		{"emojis_mixed", "hola 😊 mundo 🌍"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tuple.NewStringValue(tt.value)
			data, err := v.Serialize()
			if err != nil {
				t.Fatalf("Serialize() error = %v", err)
			}

			restored, err := tuple.NewStringValueFromBytes(data)
			if err != nil {
				t.Fatalf("NewStringValueFromBytes() error = %v", err)
			}

			if restored.Value != tt.value {
				t.Errorf("Value = %q, want %q", restored.Value, tt.value)
			}
		})
	}
}

func TestStringValueFromBytesInvalidInput(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		v, err := tuple.NewStringValueFromBytes(nil)
		if v != nil {
			t.Errorf("NewStringValueFromBytes(nil) = %v, want nil", v)
		}
		if err != tuple.ErrInvalidValue {
			t.Errorf("error = %v, want %v", err, tuple.ErrInvalidValue)
		}
	})

	t.Run("empty", func(t *testing.T) {
		v, err := tuple.NewStringValueFromBytes([]byte{})
		if v != nil {
			t.Errorf("NewStringValueFromBytes(empty) = %v, want nil", v)
		}
		if err != tuple.ErrInvalidValue {
			t.Errorf("error = %v, want %v", err, tuple.ErrInvalidValue)
		}
	})

	t.Run("too_short", func(t *testing.T) {
		v, err := tuple.NewStringValueFromBytes([]byte{0x01})
		if v != nil {
			t.Errorf("NewStringValueFromBytes(1byte) = %v, want nil", v)
		}
		if err != tuple.ErrInvalidValue {
			t.Errorf("error = %v, want %v", err, tuple.ErrInvalidValue)
		}
	})

	t.Run("length_mismatch_too_short", func(t *testing.T) {
		data := []byte{0x05, 0x00, 'h', 'e'}
		v, err := tuple.NewStringValueFromBytes(data)
		if v != nil {
			t.Errorf("NewStringValueFromBytes(length=5,data=2) = %v, want nil", v)
		}
		if err != tuple.ErrInvalidValue {
			t.Errorf("error = %v, want %v", err, tuple.ErrInvalidValue)
		}
	})

	t.Run("length_mismatch_too_long", func(t *testing.T) {
		data := []byte{0x02, 0x00, 'h', 'e', 'l', 'l', 'o'}
		v, err := tuple.NewStringValueFromBytes(data)
		if v != nil {
			t.Errorf("NewStringValueFromBytes(length=2,data=5) = %v, want nil", v)
		}
		if err != tuple.ErrInvalidValue {
			t.Errorf("error = %v, want %v", err, tuple.ErrInvalidValue)
		}
	})

	t.Run("only_length_prefix", func(t *testing.T) {
		data := []byte{0x00, 0x00}
		v, err := tuple.NewStringValueFromBytes(data)
		if err != nil {
			t.Fatalf("NewStringValueFromBytes(length=0) error = %v, want nil", err)
		}
		if v.Value != "" {
			t.Errorf("Value = %q, want empty string", v.Value)
		}
	})
}

func TestStringValueRoundTrip(t *testing.T) {
	original := tuple.NewStringValue("áéíóú über café ñoño 😊")

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	restored, err := tuple.NewStringValueFromBytes(data)
	if err != nil {
		t.Fatalf("NewStringValueFromBytes() error = %v", err)
	}

	if restored.Value != original.Value {
		t.Errorf("round trip = %q, want %q", restored.Value, original.Value)
	}
}

func TestStringValueString(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"empty", ""},
		{"ascii", "hello"},
		{"tildes", "café"},
		{"emoji", "😊"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tuple.NewStringValue(tt.value)
			got := v.String()
			if got != tt.value {
				t.Errorf("String() = %q, want %q", got, tt.value)
			}
		})
	}
}

func TestStringValueImplementsValue(t *testing.T) {
	var v tuple.Value = tuple.NewStringValue("hello")

	if v.Type() != catalog.TypeStringType {
		t.Errorf("Type() = %v, want %v", v.Type(), catalog.TypeStringType)
	}

	data, err := v.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}
	if int(v.Size()) != len(data) {
		t.Errorf("Size() = %d, len(Serialize()) = %d", v.Size(), len(data))
	}
}

func TestStringValueZeroValue(t *testing.T) {
	v := tuple.StringValue{}

	if v.Value != "" {
		t.Errorf("default Value = %q, want empty string", v.Value)
	}

	if v.Type() != catalog.TypeStringType {
		t.Errorf("Type() = %v, want %v", v.Type(), catalog.TypeStringType)
	}

	data, err := v.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	if len(data) != catalog.StringLengthSize {
		t.Errorf("len(data) = %d, want %d", len(data), catalog.StringLengthSize)
	}

	restored, err := tuple.NewStringValueFromBytes(data)
	if err != nil {
		t.Fatalf("NewStringValueFromBytes() error = %v", err)
	}
	if restored.Value != "" {
		t.Errorf("restored value = %q, want empty", restored.Value)
	}
}

func TestStringValueMaxLength(t *testing.T) {
	// Create a string of exactly MaxUint16 bytes
	s := strings.Repeat("a", math.MaxUint16)
	v := tuple.NewStringValue(s)

	data, err := v.Serialize()
	if err != nil {
		t.Fatalf("Serialize(MaxUint16) error = %v, want nil", err)
	}

	if len(data) != catalog.StringLengthSize+math.MaxUint16 {
		t.Errorf("len(data) = %d, want %d", len(data), catalog.StringLengthSize+math.MaxUint16)
	}

	restored, err := tuple.NewStringValueFromBytes(data)
	if err != nil {
		t.Fatalf("NewStringValueFromBytes() error = %v", err)
	}

	if len(restored.Value) != math.MaxUint16 {
		t.Errorf("restored len = %d, want %d", len(restored.Value), math.MaxUint16)
	}

	if restored.Value != s {
		t.Errorf("max length string mismatch on round trip")
	}
}

func TestStringValueUTF8Tildes(t *testing.T) {
	// Comprehensive UTF-8 with tildes and special chars
	inputs := []string{
		"áéíóú",
		"ÁÉÍÓÚ",
		"ñÑüÜ",
		"übersicht",
		"mañana",
		"coração",
		"façade",
		"àèìòù",
		"äëïöü",
		"âêîôû",
		"日本国",
		"中国",
		"français",
		"Déjà vu",
		"Señor",
		"¿Cómo estás?",
		"Straße",     // German ß
		"Øl i Norge", // Norwegian/Danish
		"Čeština",    // Czech
		"Русский",    // Russian
		"Ελληνικά",   // Greek
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			v := tuple.NewStringValue(input)

			data, err := v.Serialize()
			if err != nil {
				t.Fatalf("Serialize(%q) error = %v", input, err)
			}

			restored, err := tuple.NewStringValueFromBytes(data)
			if err != nil {
				t.Fatalf("NewStringValueFromBytes() error = %v", err)
			}

			if restored.Value != input {
				t.Errorf("round trip = %q, want %q", restored.Value, input)
			}
		})
	}
}
