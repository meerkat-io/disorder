package loader

import "encoding/json"

type jsonUnmarshaller struct {
}

func (*jsonUnmarshaller) Unmarshal(data []byte, schema *schemaFile) error {
	return json.Unmarshal(data, schema)
}

func NewJsonLoader() Loader {
	return newLoaderImpl(&jsonUnmarshaller{})
}
