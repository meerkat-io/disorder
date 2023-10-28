package disorder

type Enum interface {
	Enum()
	GetValue() (string, error)
	SetValue(enum string) error
}
