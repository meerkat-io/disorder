package loader_test

import (
	"fmt"
	"testing"

	"github.com/meerkat-lib/disorder/internal/generator"
	"github.com/meerkat-lib/disorder/internal/loader"
)

func TestLoadYamlFile(t *testing.T) {
	loader := loader.NewYamlLoader()
	files, err := loader.Load("./test_data/schema.yaml")
	fmt.Println(err)
	generator := generator.NewGoGenerator()
	_ = generator.Generate(".", files)
	t.Fail()
}

/*
func TestLoadJsonFile(t *testing.T) {
	loader := loader.NewJsonLoader()
	_, err := loader.Load("./test_data/schema.json")
	fmt.Println(err)
	t.Fail()
}*/
