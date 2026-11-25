// Package gdscript generates GDScript (Godot 4) classes from schema files.
package gdscript

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/meerkat-io/bloom/folder"
	"github.com/meerkat-io/bloom/strcase"
	"github.com/meerkat-io/disorder/internal/generator"
	"github.com/meerkat-io/disorder/internal/schema"
)

// classPrefix is the schema option holding a prefix for all generated class
// and enum names, used to avoid clashes with Godot built-in names.
const classPrefix = "gdscript_class_prefix"

// builtinNames are Godot names that generated classes and enums must not
// shadow. A clash makes the generated script fail to compile.
var builtinNames = map[string]bool{
	"Object": true, "RefCounted": true, "Resource": true, "Node": true, "Variant": true,
	"Color": true, "Vector2": true, "Vector2i": true, "Vector3": true, "Vector3i": true,
	"Vector4": true, "Vector4i": true, "Rect2": true, "Rect2i": true, "Transform2D": true,
	"Transform3D": true, "Basis": true, "Plane": true, "Quaternion": true, "AABB": true,
	"Projection": true, "Callable": true, "Signal": true, "Dictionary": true, "Array": true,
	"String": true, "StringName": true, "NodePath": true, "RID": true, "Time": true,
	"JSON": true, "Image": true, "Texture": true, "Timer": true, "Thread": true,
}

// NewGenerator returns a generator that writes GDScript, one file per schema.
func NewGenerator() generator.Generator {
	return &gdGenerator{
		template: template.Must(template.New("gdscript").Parse(fileTemplate)),
	}
}

type gdGenerator struct {
	template *template.Template
}

type fileView struct {
	Imports  []importView
	Enums    []enumView
	Messages []messageView
}

type importView struct {
	Alias string
	Path  string
}

type enumView struct {
	Name        string
	ValuesConst string
	Members     []string
	Values      []string // quoted wire values
}

type messageView struct {
	Name   string
	Fields []fieldView
}

type fieldView struct {
	Name    string
	Type    string
	Default string
	Desc    string
	Comment string
}

// builder holds what is needed to turn the types of one file into GDScript.
type builder struct {
	file          *schema.File
	files         map[string]*schema.File
	qualifiedPath map[string]string
	imports       map[string]importView // by target file path
}

func (g *gdGenerator) Generate(dir string, files map[string]*schema.File, qualifiedPath map[string]string) error {
	for _, file := range files {
		if err := g.generateFile(dir, file, files, qualifiedPath); err != nil {
			return fmt.Errorf("generate %s failed: %w", file.FilePath, err)
		}
	}
	return nil
}

func (g *gdGenerator) generateFile(dir string, file *schema.File, files map[string]*schema.File, qualifiedPath map[string]string) error {
	b := &builder{file: file, files: files, qualifiedPath: qualifiedPath, imports: map[string]importView{}}
	view, err := b.build()
	if err != nil {
		return err
	}
	buf := &bytes.Buffer{}
	if err := g.template.Execute(buf, view); err != nil {
		return err
	}

	schemaDir, err := filepath.Abs(filepath.Join(dir, filepath.FromSlash(packageFolder(file.Package))))
	if err != nil {
		return err
	}
	if err := folder.Create(schemaDir); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(schemaDir, outputName(file)), buf.Bytes(), 0666)
}

func (b *builder) build() (*fileView, error) {
	view := &fileView{}
	prefix := b.file.Options[classPrefix]
	for _, enum := range b.file.Enums {
		name, err := className(prefix, enum.Name)
		if err != nil {
			return nil, err
		}
		members := map[string]bool{}
		e := enumView{Name: name, ValuesConst: valuesConst(prefix, enum.Name)}
		for _, value := range enum.Values {
			member := strings.ToUpper(strcase.SnakeCase(value))
			if members[member] {
				return nil, fmt.Errorf("enum %s: values clash as GDScript constant %s", enum.Name, member)
			}
			members[member] = true
			e.Members = append(e.Members, member)
			e.Values = append(e.Values, strconv.Quote(value))
		}
		view.Enums = append(view.Enums, e)
	}
	for _, message := range b.file.Messages {
		name, err := className(prefix, message.Name)
		if err != nil {
			return nil, err
		}
		m := messageView{Name: name}
		for _, field := range message.Fields {
			f, err := b.field(field)
			if err != nil {
				return nil, fmt.Errorf("message %s: %w", message.Name, err)
			}
			m.Fields = append(m.Fields, f)
		}
		view.Messages = append(view.Messages, m)
	}
	for _, imp := range b.imports {
		view.Imports = append(view.Imports, imp)
	}
	sort.Slice(view.Imports, func(i, j int) bool { return view.Imports[i].Alias < view.Imports[j].Alias })
	return view, nil
}

func (b *builder) field(field *schema.Field) (fieldView, error) {
	t, err := b.typeOf(field.Type)
	if err != nil {
		return fieldView{}, err
	}
	f := fieldView{Name: field.Name, Type: t.decl, Default: t.value, Desc: t.desc}
	if field.Type.Type == schema.TypeTimestamp {
		f.Comment = "unix time in milliseconds"
	}
	return f, nil
}

// gdType is a type as seen by GDScript: the declaration, its default value and
// the descriptor read by the codec.
type gdType struct {
	decl  string
	value string
	desc  string
}

func (b *builder) typeOf(info *schema.TypeInfo) (gdType, error) {
	switch info.Type {
	case schema.TypeBool:
		return gdType{"bool", "false", `"bool"`}, nil
	case schema.TypeInt:
		return gdType{"int", "0", `"int"`}, nil
	case schema.TypeLong:
		return gdType{"int", "0", `"long"`}, nil
	case schema.TypeFloat:
		return gdType{"float", "0.0", `"float"`}, nil
	case schema.TypeDouble:
		return gdType{"float", "0.0", `"double"`}, nil
	case schema.TypeString:
		return gdType{"String", `""`, `"string"`}, nil
	case schema.TypeBytes:
		return gdType{"PackedByteArray", "PackedByteArray()", `"bytes"`}, nil
	case schema.TypeTimestamp:
		return gdType{"int", "0", `"timestamp"`}, nil
	case schema.TypeEnum:
		target, scope, name, err := b.target(info)
		if err != nil {
			return gdType{}, err
		}
		enum := findEnum(target, name)
		first := strings.ToUpper(strcase.SnakeCase(enum.Values[0]))
		typ, _ := className(target.Options[classPrefix], name)
		return gdType{
			decl:  scope + typ,
			value: scope + typ + "." + first,
			desc:  fmt.Sprintf(`["enum", %s%s]`, scope, valuesConst(target.Options[classPrefix], name)),
		}, nil
	case schema.TypeObject:
		target, scope, name, err := b.target(info)
		if err != nil {
			return gdType{}, err
		}
		typ, _ := className(target.Options[classPrefix], name)
		return gdType{decl: scope + typ, value: "null", desc: fmt.Sprintf(`["object", %s%s]`, scope, typ)}, nil
	case schema.TypeArray:
		element, err := b.typeOf(info.ElementType)
		if err != nil {
			return gdType{}, err
		}
		decl := "Array"
		// Godot 4 has no nested typed arrays.
		if !isContainer(info.ElementType) {
			decl = fmt.Sprintf("Array[%s]", arrayElementType(info.ElementType, element))
		}
		return gdType{decl, "[]", fmt.Sprintf(`["array", %s]`, element.desc)}, nil
	case schema.TypeMap:
		element, err := b.typeOf(info.ElementType)
		if err != nil {
			return gdType{}, err
		}
		return gdType{"Dictionary", "{}", fmt.Sprintf(`["map", %s]`, element.desc)}, nil
	}
	return gdType{}, fmt.Errorf("unsupported type %q", info.TypeRef)
}

// arrayElementType is the element type used in Array[...]. Enums are stored as
// plain ints, which typed arrays accept in every Godot 4 version.
func arrayElementType(info *schema.TypeInfo, element gdType) string {
	if info.Type == schema.TypeEnum {
		return "int"
	}
	return element.decl
}

func isContainer(info *schema.TypeInfo) bool {
	return info.Type == schema.TypeArray || info.Type == schema.TypeMap
}

// target finds the file defining a referenced type, and the expression that
// scopes it: empty for the current file, otherwise the alias of a preloaded
// const.
func (b *builder) target(info *schema.TypeInfo) (target *schema.File, scope, name string, err error) {
	targetPath, ok := b.qualifiedPath[info.Qualified]
	if !ok {
		return nil, "", "", fmt.Errorf("unresolved type %q", info.TypeRef)
	}
	target = b.files[targetPath]
	name = info.Qualified[strings.LastIndex(info.Qualified, ".")+1:]
	if target.FilePath == b.file.FilePath {
		return target, "", name, nil
	}
	imp, ok := b.imports[target.FilePath]
	if !ok {
		rel, err := filepath.Rel(filepath.FromSlash(packageFolder(b.file.Package)), filepath.Join(filepath.FromSlash(packageFolder(target.Package)), outputName(target)))
		if err != nil {
			return nil, "", "", err
		}
		imp = importView{Alias: importAlias(target), Path: filepath.ToSlash(rel)}
		b.imports[target.FilePath] = imp
	}
	return target, imp.Alias + ".", name, nil
}

func findEnum(file *schema.File, name string) *schema.Enum {
	for _, enum := range file.Enums {
		if enum.Name == name {
			return enum
		}
	}
	return nil
}

// className returns the GDScript name of a message or enum.
func className(prefix, name string) (string, error) {
	full := prefix + strcase.PascalCase(name)
	if builtinNames[full] {
		return "", fmt.Errorf("%q clashes with a Godot built-in name, set option %s to add a prefix", full, classPrefix)
	}
	return full, nil
}

func valuesConst(prefix, name string) string {
	return strings.ToUpper(strcase.SnakeCase(prefix+strcase.PascalCase(name))) + "_VALUES"
}

func importAlias(file *schema.File) string {
	base := strings.TrimSuffix(filepath.Base(file.FilePath), filepath.Ext(file.FilePath))
	return "_" + strcase.PascalCase(strings.ReplaceAll(file.Package, ".", "_")+"_"+base)
}

func outputName(file *schema.File) string {
	return strings.TrimSuffix(filepath.Base(file.FilePath), filepath.Ext(file.FilePath)) + ".gd"
}

func packageFolder(pkg string) string {
	folders := strings.Split(pkg, ".")
	for i := range folders {
		folders[i] = strcase.SnakeCase(folders[i])
	}
	return path.Join(folders...)
}
