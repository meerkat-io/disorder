package golang

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/meerkat-io/bloom/folder"
	"github.com/meerkat-io/bloom/format/strcase"
	"github.com/meerkat-io/disorder/internal/generator"
	"github.com/meerkat-io/disorder/internal/schema"
)

const (
	golang          = "golang"
	goPackagePrefix = "go_package_prefix"
)

func NewGoGenerator() generator.Generator {
	g := &goGenerator{}
	g.initTemplete()
	return g
}

type goGenerator struct {
	define *template.Template
	rpc    *template.Template
}

type goSchema struct {
	Schema        *schema.File
	DefineImports []string
	RpcImports    []string
}

func (g *goGenerator) Generate(dir string, files map[string]*schema.File, qualifiedPath map[string]string) error {
	for _, file := range files {
		schemaFile := &goSchema{
			Schema: file,
		}

		defineImports := make(map[string]bool)
		rpcImports := make(map[string]bool)
		for _, message := range file.Messages {
			for _, field := range message.Fields {
				g.resolveImport(field.Type, defineImports, file, files, qualifiedPath)
			}
		}
		for _, service := range file.Services {
			for _, rpc := range service.Rpc {
				g.resolveImport(rpc.Input, defineImports, file, files, qualifiedPath)
				g.resolveImport(rpc.Output, defineImports, file, files, qualifiedPath)
			}
		}

		if len(file.Enums) > 0 {
			defineImports["fmt"] = true
		}
		rpcImports["fmt"] = true
		rpcImports["github.com/meerkat-io/disorder"] = true
		rpcImports["github.com/meerkat-io/disorder/rpc"] = true
		rpcImports["github.com/meerkat-io/disorder/rpc/code"] = true
		for path := range defineImports {
			schemaFile.DefineImports = append(schemaFile.DefineImports, path)
		}
		for path := range rpcImports {
			schemaFile.RpcImports = append(schemaFile.RpcImports, path)
		}

		schemaDir, err := filepath.Abs(filepath.Join(dir, g.packageFolder(file.Package)))
		if err != nil {
			return err
		}
		err = folder.Create(schemaDir)
		if err != nil {
			return err
		}

		buf := &bytes.Buffer{}
		if err := g.define.Execute(buf, schemaFile); err != nil {
			return err
		}
		source := buf.Bytes()
		if source, err = format.Source(source); err != nil {
			return err
		}
		schemaFilePath := filepath.Base(file.FilePath)
		schemaFilePath = fmt.Sprintf("%s.go", strings.TrimSuffix(schemaFilePath, filepath.Ext(schemaFilePath)))
		err = os.WriteFile(filepath.Join(schemaDir, schemaFilePath), source, 0666)
		if err != nil {
			return err
		}
		/*
			if len(file.Services) > 0 {
				buf = &bytes.Buffer{}
				if err := g.rpc.Execute(buf, file); err != nil {
					return err
				}
				source = buf.Bytes()
				if source, err = format.Source(source); err != nil {
					return err
				}
				schemaFile = filepath.Base(file.FilePath)
				schemaFile = fmt.Sprintf("%s_rpc.go", strings.TrimSuffix(schemaFile, filepath.Ext(schemaFile)))
				err = os.WriteFile(filepath.Join(schemaDir, schemaFile), source, 0666)
				if err != nil {
					return err
				}
			}*/
	}
	return nil
}

func (g *goGenerator) packageFolder(pkg string) string {
	folders := strings.Split(pkg, ".")
	for i := range folders {
		folders[i] = strcase.SnakeCase(folders[i])
	}
	return strings.Join(folders, "/")
}

func (g *goGenerator) resolveImport(typeInfo *schema.TypeInfo, importMap map[string]bool, current *schema.File, files map[string]*schema.File, qualifiedPath map[string]string) {
	if typeInfo.Type == schema.TypeTimestamp {
		importMap["time"] = true
	}
	if typeInfo.Qualified != "" && current.FilePath != qualifiedPath[typeInfo.Qualified] {
		targetFile := files[qualifiedPath[typeInfo.Qualified]]
		prefix := ""
		if targetFile.Options != nil {
			prefix = targetFile.Options[goPackagePrefix]
		}
		importMap[filepath.Join(prefix, g.packageFolder(targetFile.Package))] = true
	}
}
