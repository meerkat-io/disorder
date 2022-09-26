package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/meerkat-io/bloom/flag"
	"github.com/meerkat-io/bloom/folder"

	"github.com/meerkat-io/disorder/internal/generator/golang"
	"github.com/meerkat-io/disorder/internal/loader"
)

type flags struct {
	Help bool `flag:"h" usage:"help"`
	//Lang   string `flag:"t" usage:"language template" tip:"go|cs|java" required:"true"`
	Input  string `flag:"i" usage:"input schema (yaml, json or toml)" tip:"input" required:"true"`
	Output string `flag:"o" usage:"output folder" tip:"output" default:"."`
}

func main() {
	f := &flags{}
	err := flag.ParseCommandLine(f)
	if f.Help {
		flag.Usage()
		os.Exit(0)
	}
	if err != nil {
		flag.Usage()
		fmt.Println(err)
		os.Exit(0)
	}

	if exists := folder.Exists(f.Output); !exists {
		fmt.Printf("folder \"%s\" does not exist\n", f.Output)
		os.Exit(0)
	}

	var l loader.Loader
	ext := filepath.Ext(f.Input)
	if ext == ".yaml" {
		l = loader.NewYamlLoader()
	} else if ext == ".json" {
		l = loader.NewJsonLoader()
	} else if ext == ".toml" {
		l = loader.NewTomlLoader()
	} else {
		fmt.Println("unknow schema type. should be yaml|json|toml file")
		os.Exit(0)
	}
	files, qualifiedPath, err := l.Load(f.Input)
	if err != nil {
		fmt.Printf("load file %s failed: %s", f.Input, err.Error())
		os.Exit(0)
	}

	//TO-DO check language
	generator := golang.NewGoGenerator()
	err = generator.Generate(f.Output, files, qualifiedPath)
	if err != nil {
		fmt.Printf("generate failed: %s", err.Error())
	}
}
