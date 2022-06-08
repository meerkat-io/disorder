package disorder_test

import (
	"fmt"
	"testing"

	"github.com/meerkat-lib/disorder"
	"github.com/meerkat-lib/disorder/internal/generator"
	"github.com/meerkat-lib/disorder/internal/loader"
	"github.com/meerkat-lib/disorder/internal/test_data/test"
)

func TestLoadYamlFile(t *testing.T) {
	loader := loader.NewYamlLoader()
	files, err := loader.Load("./internal/test_data/schema.yaml")
	fmt.Println(err)
	generator := generator.NewGoGenerator()
	err = generator.Generate("./internal", files)
	fmt.Println(err)
	t.Fail()
}

/*
func TestLoadJsonFile(t *testing.T) {
	loader := loader.NewJsonLoader()
	_, err := loader.Load("./test_data/schema.json")
	fmt.Println(err)
	t.Fail()
}*/

func TestMarshal(t *testing.T) {
	cat := test.AnimalCat
	data, err := disorder.Marshal(&cat)
	fmt.Println(data)
	fmt.Println(err)
	t.Fail()
}
