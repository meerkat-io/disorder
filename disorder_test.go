package disorder_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/meerkat-lib/disorder"
	"github.com/meerkat-lib/disorder/internal/generator/golang"
	"github.com/meerkat-lib/disorder/internal/loader"
	"github.com/meerkat-lib/disorder/internal/test_data/test"
	"github.com/meerkat-lib/disorder/internal/test_data/test/sub"
	"github.com/meerkat-lib/disorder/rpc"
)

type testService struct {
}

func (*testService) Increase(c *rpc.Context, request int32) (int32, *rpc.Error) {
	request++
	return request, nil
}

func (*testService) GetAnotherObject(c *rpc.Context, request string) (*test.AnotherObject, *rpc.Error) {
	fmt.Println(request)
	return &test.AnotherObject{
		Value: 456,
	}, nil
}

func (*testService) PrintSubObject(c *rpc.Context, request *sub.SubObject) (int32, *rpc.Error) {
	fmt.Printf("%v\n", *request)
	return 123, nil
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

	input := map[string]string{
		"foo": "bar",
	}
	data, err := disorder.Marshal(input)
	fmt.Println(err)
	fmt.Println(input)
	fmt.Println(data)
	output := make(map[string]string)
	err = disorder.Unmarshal(data, &output)
	fmt.Println(err)
	fmt.Println(output)

	t.Fail()
}

func TestRpc(t *testing.T) {
	s := rpc.NewServer()
	sub.RegisterMathServiceServer(s, &testService{})
	_ = s.Listen(":8888")

	time.Sleep(time.Second)

	c := sub.NewMathServiceClient(rpc.NewClient("localhost:8888"))
	result, rpcErr := c.Increase(rpc.NewContext(), 17)
	fmt.Println(rpcErr)
	fmt.Println(result)

	t.Fail()
}

func TestRpc2(t *testing.T) {
	s := rpc.NewServer()
	test.RegisterPrimaryServiceServer(s, &testService{})
	_ = s.Listen(":8888")

	time.Sleep(time.Second)

	c := test.NewPrimaryServiceClient(rpc.NewClient("localhost:8888"))

	result1, rpcErr := c.GetAnotherObject(rpc.NewContext(), "foo.bar")
	fmt.Println(rpcErr)
	fmt.Println(*result1)

	result2, rpcErr := c.PrintSubObject(rpc.NewContext(), &sub.SubObject{
		Value: 789,
	})
	fmt.Println(rpcErr)
	fmt.Println(result2)

	t.Fail()
}
