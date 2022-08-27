package rpc

type title string

func (t *title) Enum() {}
func (t *title) Decode(enum string) error {
	*t = title(enum)
	return nil
}
func (t *title) Encode() (string, error) {
	return string(*t), nil
}
