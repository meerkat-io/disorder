package disorder_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/meerkat-lib/disorder"
	"github.com/meerkat-lib/disorder/internal/generator/golang"
	"github.com/meerkat-lib/disorder/internal/loader"
	"github.com/meerkat-lib/disorder/internal/test_data/test/sub"
	"github.com/meerkat-lib/disorder/rpc"
	"github.com/meerkat-lib/disorder/rpc/code"
)

type mathService struct {
}

func (*mathService) Increase(c *rpc.Context, request *int32) (*int32, *rpc.Error) {
	return nil, &rpc.Error{
		Code:  code.Internal,
		Error: fmt.Errorf("caculator broken"),
	}
	/*
		fmt.Println("increase from server")
		value := *request
		value++
		fmt.Println(value)
		return &value, nil*/
}

func TestLoadYamlFile(t *testing.T) {
	loader := loader.NewYamlLoader()
	files, err := loader.Load("./internal/test_data/schema.yaml")
	fmt.Println(err)
	generator := golang.NewGoGenerator()
	err = generator.Generate("./internal", files)
	fmt.Println(err)
	t.Fail()
}

func TestLoadJsonFile(t *testing.T) {
	loader := loader.NewJsonLoader()
	files, err := loader.Load("./internal/test_data/schema.json")
	fmt.Println(err)
	generator := golang.NewGoGenerator()
	err = generator.Generate("./internal", files)
	fmt.Println(err)
	t.Fail()
}

func TestLoadTomlFile(t *testing.T) {
	loader := loader.NewTomlLoader()
	files, err := loader.Load("./internal/test_data/schema.toml")
	fmt.Println(err)
	generator := golang.NewGoGenerator()
	err = generator.Generate("./internal", files)
	fmt.Println(err)
	t.Fail()
}

func TestMarshal(t *testing.T) {

	input := map[string]string{}
	data, err := disorder.Marshal(input)
	fmt.Println(err)
	fmt.Println(input)
	fmt.Println(data)
	var output interface{}
	err = disorder.Unmarshal(data, &output)
	fmt.Println(err)
	fmt.Println(output)

	t.Fail()
}

func TestRpc(t *testing.T) {
	s := rpc.NewServer()
	sub.RegisterMathService(s, &mathService{})
	_ = s.Listen(":8888")

	time.Sleep(time.Second)

	c := sub.NewMathServiceClient(rpc.NewClient("localhost:8888"))
	value := int32(10)
	result, rpcErr := c.Increase(rpc.NewContext(), &value)
	fmt.Println(rpcErr)
	fmt.Println(*result)

	t.Fail()
}
