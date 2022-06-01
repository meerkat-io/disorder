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
