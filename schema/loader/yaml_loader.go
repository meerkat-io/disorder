package loader

import "gopkg.in/yaml.v3"

type yamlUnmarshaller struct {
}

func (*yamlUnmarshaller) Unmarshal(data []byte, schema *schemaFile) error {
	return yaml.Unmarshal(data, schema)
}

func NewYamlLoader() Loader {
	return newLoaderImpl(&yamlUnmarshaller{})
}
