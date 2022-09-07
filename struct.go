package disorder

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

var (
	structsMap      = make(map[reflect.Type]*structInfo)
	structsMapMutex sync.RWMutex
)

type structInfo struct {
	fieldsMap  map[string]*fieldInfo
	fieldsList []*fieldInfo
}

type fieldInfo struct {
	key   string
	index int
}

func getStructInfo(typ reflect.Type) (*structInfo, error) {
	structsMapMutex.RLock()
	s, exists := structsMap[typ]
	structsMapMutex.RUnlock()
	if exists {
		return s, nil
	}

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
		if len(fields) > 0 {
			tag = fields[0]
			if len(fields) > 1 {
				return nil, fmt.Errorf("unsupported flag %s in tag %s", fields[1], tag)
			}
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

func isNull(value reflect.Value) bool {
	if !value.IsValid() {
		return true
	}
	switch value.Kind() {
	case reflect.Interface, reflect.Ptr, reflect.Slice, reflect.Map:
		if value.IsNil() {
			return true
		}
	}
	return false
}
