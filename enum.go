package disorder

// Enum is implemented by generated enum types. Values are strings, validated
// by SetValue and GetValue.
type Enum interface {
	Enum()
	GetValue() (string, error)
	SetValue(enum string) error
}
