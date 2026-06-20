package catalog

type DataType uint8

const (
	TypeInt DataType = iota
	TypeString
	TypeBool
)

func (d DataType) IsValid() bool {
	return d >= TypeInt && d <= TypeBool
}
