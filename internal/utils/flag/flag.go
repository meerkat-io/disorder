package flag

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"strconv"
	"unsafe"
)

/**********************************************************
type Flags struct {
	Help    bool   `flag:"h" usage:"help"`
	Lang    string `flag:"t" usage:"language template" tip:"go|cs" required:"true"`
	Input   string `flag:"i" usage:"input file or folder" tip:"input" default:"."`
	Output  string `flag:"o" usage:"output folder" tip:"output" default:"."`
	Package string `flag:"p" usage:"schema package overwrite" tip:"package"`
}
// Supported types: int, string, float64, bool
***********************************************************/

var flags map[string]*flag

type flag struct {
	Name     string // name as it appears on command line
	Usage    string // help message
	Tip      string // short help message
	Default  string // default value (as text); for usage message
	Required bool   // if required value

	Value   value
	Visited bool
}

type value interface {
	Get() interface{}
	Set(string) error
	IsBool() bool
}

type boolValue bool

func newBoolValue(val bool, p *bool) *boolValue {
	*p = val
	return (*boolValue)(p)
}

func (b *boolValue) Get() interface{} { return bool(*b) }

func (b *boolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	*b = boolValue(v)
	return err
}

func (b *boolValue) IsBool() bool { return true }

type stringValue string

func newStringValue(val string, p *string) *stringValue {
	*p = val
	return (*stringValue)(p)
}

func (s *stringValue) Get() interface{} { return string(*s) }

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) IsBool() bool { return false }

type intValue int

func newIntValue(val int, p *int) *intValue {
	*p = val
	return (*intValue)(p)
}

func (i *intValue) Get() interface{} { return int(*i) }

func (i *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, strconv.IntSize)
	*i = intValue(v)
	return err
}

func (i *intValue) IsBool() bool { return false }

type floatValue float64

func newFloatValue(val float64, p *float64) *floatValue {
	*p = val
	return (*floatValue)(p)
}

func (f *floatValue) Get() interface{} { return float64(*f) }

func (f *floatValue) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	*f = floatValue(v)
	return err
}

func (f *floatValue) IsBool() bool { return false }

// Parse flag
func Parse(config interface{}) error {
	flags = make(map[string]*flag)

	info := reflect.TypeOf(config)
	if info.Kind() != reflect.Ptr {
		return fmt.Errorf("must use pointer of config struct")
	}
	info = info.Elem()

	value := reflect.ValueOf(config)
	value = value.Elem()

	var err error
	num := info.NumField()
	for i := 0; i < num; i++ {
		f := &flag{}
		field := info.Field(i)

		f.Name = field.Tag.Get("flag")
		if f.Name == "" {
			return fmt.Errorf("flag is empty: %s", field.Name)
		}
		f.Default = field.Tag.Get("default")
		f.Usage = field.Tag.Get("usage")
		f.Tip = field.Tag.Get("tip")
		f.Required = field.Tag.Get("required") == "true"

		switch value.Field(i).Interface().(type) {
		case bool:
			f.Value = newBoolValue(f.Default == "true", (*bool)(unsafe.Pointer(value.Field(i).Addr().Pointer())))
		case string:
			f.Value = newStringValue(f.Default, (*string)(unsafe.Pointer(value.Field(i).Addr().Pointer())))
		case int:
			var v int64
			if f.Default != "" {
				v, err = strconv.ParseInt(f.Default, 10, 32)
				if err != nil {
					return fmt.Errorf("invalid flag default value: [name:%s], [value:%s]", field.Name, f.Default)
				}
			}
			f.Value = newIntValue(int(v), (*int)(unsafe.Pointer(value.Field(i).Addr().Pointer())))
		case float64:
			var v float64
			if f.Default != "" {
				v, err = strconv.ParseFloat(f.Default, 64)
				if err != nil {
					return fmt.Errorf("invalid flag default value: [name:%s], [value:%s]", field.Name, f.Default)
				}
			}
			f.Value = newFloatValue(v, (*float64)(unsafe.Pointer(value.Field(i).Addr().Pointer())))
		default:
			return fmt.Errorf("invalid flag type: [name:%s], [value:%s]", field.Name, field.Type.String())
		}

		flags[f.Name] = f
	}

	args := os.Args[1:]

	for len(args) > 0 {
		s := args[0]
		if len(s) < 2 || s[0] != '-' {
			return fmt.Errorf("invalid flag %s", s)
		}
		name := s[1:]
		if name[0] == '-' {
			return fmt.Errorf("bad flag syntax: %s", s)
		}
		f, ok := flags[name]
		if !ok {
			return fmt.Errorf("flag provided but not defined: -%s", name)
		}
		f.Visited = true
		args = args[1:]
		if f.Value.IsBool() {
			_ = f.Value.Set("true")
		} else {
			if len(args) == 0 || args[0][0] == '-' {
				return fmt.Errorf("flag -%s has no value", name)
			}
			err = f.Value.Set(args[0])
			if err != nil {
				return err
			}
			args = args[1:]
		}
	}

	for _, f := range flags {
		if !f.Visited && f.Required {
			return fmt.Errorf("flag -%s is required but not set", f.Name)
		}
	}

	return nil
}

// Usage print usage
func Usage() {
	file := path.Base(os.Args[0])
	fmt.Printf("\nUsage: %s ", file)
	for _, f := range flags {
		fmt.Printf("[-%s %s]", f.Name, f.Tip)
	}
	fmt.Print("\n\nOptions:\n")
	for _, f := range flags {
		fmt.Printf("    -%s:    %s\n", f.Name, f.Usage)
	}
	fmt.Println("")
}
