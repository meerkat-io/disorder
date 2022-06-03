package loader_test

import (
	"fmt"
	"testing"

	"github.com/meerkat-lib/disorder/schema/loader"
)

func TestLoadYamlFile(t *testing.T) {
	loader := loader.NewYamlLoader()
	_, err := loader.Load("../../test_data/schema.yaml")
	fmt.Println(err)
	t.Fail()
}

func TestLoadJsonFile(t *testing.T) {
	loader := loader.NewJsonLoader()
	_, err := loader.Load("../../test_data/schema.json")
	fmt.Println(err)
	t.Fail()
}
