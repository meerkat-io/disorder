package disorder_test

import (
	"fmt"
	"testing"

	"github.com/meerkat-lib/disorder"
	"github.com/meerkat-lib/disorder/internal/generator/golang"
	"github.com/meerkat-lib/disorder/internal/loader"
	"github.com/meerkat-lib/disorder/internal/test_data/test"
)

func TestLoadYamlFile(t *testing.T) {
	loader := loader.NewYamlLoader()
	files, err := loader.Load("./internal/test_data/schema.yaml")
	fmt.Println(err)
	generator := golang.NewGoGenerator()
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
	enum := test.AnimalDog
	var input disorder.Enum = &enum
	data, err := disorder.Marshal(input)
	fmt.Println(err)
	fmt.Println(data)
	var output interface{}
	err = disorder.Unmarshal(data, &output)
	fmt.Println(err)
	fmt.Println(output)
	t.Fail()
}
