package rpc

type title string

func (t *title) Enum() {}
func (t *title) FromString(enum string) error {
	*t = title(enum)
	return nil
}
func (t *title) ToString() (string, error) {
	return string(*t), nil
}
