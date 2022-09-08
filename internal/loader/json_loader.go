package loader

import "encoding/json"

func NewJsonLoader() Loader {
	return newLoader(&jsonUnmarshaller{})
}

type jsonUnmarshaller struct {
}

func (*jsonUnmarshaller) unmarshal(data []byte, schema *proto) error {
	return json.Unmarshal(data, schema)
}
