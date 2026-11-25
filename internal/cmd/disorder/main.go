package main

import (
	"fmt"
	"os"

	"github.com/meerkat-io/bloom/flag"
	"github.com/meerkat-io/bloom/folder"

	"github.com/meerkat-io/disorder/internal/generator"
	"github.com/meerkat-io/disorder/internal/generator/gdscript"
	"github.com/meerkat-io/disorder/internal/generator/golang"
	"github.com/meerkat-io/disorder/internal/loader"
)

type flags struct {
	Help   bool   `flag:"h" usage:"help"`
	Lang   string `flag:"l" usage:"target language" tip:"go|gdscript" default:"go"`
	Input  string `flag:"i" usage:"input schema (yaml)" tip:"input" required:"true"`
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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if exists := folder.Exists(f.Output); !exists {
		fmt.Fprintf(os.Stderr, "folder \"%s\" does not exist\n", f.Output)
		os.Exit(1)
	}

	files, qualifiedPath, err := loader.NewLoader().Load(f.Input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load file %s failed: %s\n", f.Input, err.Error())
		os.Exit(1)
	}

	var gen generator.Generator
	switch f.Lang {
	case "go":
		gen = golang.NewGoGenerator()
	case "gdscript":
		gen = gdscript.NewGenerator()
	default:
		fmt.Fprintf(os.Stderr, "unsupported language \"%s\", use go or gdscript\n", f.Lang)
		os.Exit(1)
	}
	err = gen.Generate(f.Output, files, qualifiedPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate failed: %s\n", err.Error())
		os.Exit(1)
	}
}
