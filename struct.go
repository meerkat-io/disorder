package disorder

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type IsZeroer interface {
	IsZero() bool
}

var (
	structsMap      = make(map[reflect.Type]*structInfo)
	structsMapMutex sync.RWMutex
)

type structInfo struct {
	fieldsMap  map[string]*fieldInfo
	fieldsList []*fieldInfo
}

type fieldInfo struct {
	key       string
	index     int
	omitEmpty bool
}

func getStructInfo(typ reflect.Type) (*structInfo, error) {
	structsMapMutex.RLock()
	s, exists := structsMap[typ]
	structsMapMutex.RUnlock()
	if exists {
		fmt.Println("get struct info from cache")
		return s, nil
	}

	fmt.Println("create new struct info")
	count := typ.NumField()
	s = &structInfo{
		fieldsMap:  map[string]*fieldInfo{},
		fieldsList: make([]*fieldInfo, 0, count),
	}
	for i := 0; i < count; i++ {
		field := typ.Field(i)
		if field.PkgPath != "" && !field.Anonymous {
			continue // Private field
		}

		f := &fieldInfo{
			index: i,
		}
		tag := field.Tag.Get("disorder")
		if tag == "" && !strings.Contains(string(field.Tag), ":") {
			tag = string(field.Tag)
		}
		if tag == "-" {
			continue
		}
		fields := strings.Split(tag, ",")
		if len(fields) > 1 {
			for _, flag := range fields[1:] {
				switch flag {
				case "omitempty":
					f.omitEmpty = true
				default:
					return nil, fmt.Errorf("unsupported flag %s in tag %s", flag, tag)
				}
			}
			tag = fields[0]
		}
		if tag != "" {
			f.key = tag
		} else {
			f.key = field.Name
		}

		if _, exists = s.fieldsMap[f.key]; exists {
			return nil, fmt.Errorf("duplicated key '" + f.key + "' in struct " + typ.String())
		}

		s.fieldsList = append(s.fieldsList, f)
		s.fieldsMap[f.key] = f
	}
	structsMap[typ] = s
	return s, nil
}

func isZero(value reflect.Value) bool {
	kind := value.Kind()
	if z, ok := value.Interface().(IsZeroer); ok {
		if (kind == reflect.Ptr || kind == reflect.Interface) && value.IsNil() {
			return true
		}
		return z.IsZero()
	}
	switch kind {
	case reflect.String:
		return len(value.String()) == 0

	case reflect.Interface, reflect.Ptr:
		return value.IsNil()

	case reflect.Slice:
		return value.Len() == 0

	case reflect.Map:
		return value.Len() == 0

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint() == 0

	case reflect.Float32, reflect.Float64:
		return value.Float() == 0

	case reflect.Bool:
		return !value.Bool()

	case reflect.Struct:
		typ := value.Type()
		for i := value.NumField() - 1; i >= 0; i-- {
			if typ.Field(i).PkgPath != "" {
				continue // Private field
			}
			if !isZero(value.Field(i)) {
				return false
			}
		}
		return true
	}
	return false
}
