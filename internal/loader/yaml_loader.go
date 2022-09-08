package loader

import "gopkg.in/yaml.v3"

func NewYamlLoader() Loader {
	return newLoader(&yamlUnmarshaller{})
}

type yamlUnmarshaller struct {
}

func (*yamlUnmarshaller) unmarshal(data []byte, schema *proto) error {
	return yaml.Unmarshal(data, schema)
}
