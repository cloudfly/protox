package doc

import (
	"fmt"
	"reflect"
	"strings"
)

// ServiceInfo represents information about an interface.
type ServiceInfo struct {
	Name    string
	Package string
	Methods []MethodInfo
}

// MethodInfo represents a method definition in an interface.
type MethodInfo struct {
	Name   string
	Input  *TypeInfo // Input parameter type information
	Output *TypeInfo // Return value type information
}

// TypeInfo represents information about a data type.
type TypeInfo struct {
	Name   string
	Fields []FieldInfo
}

// FieldInfo represents a field in a struct.
type FieldInfo struct {
	Name string
	Type string
	Tag  string
}

// ParseService parses method information from an interface type.
// Note: input must be a pointer to an interface type (e.g., (*MyInterface)(nil))
// or a reflect.Type (more recommended), but for API simplification, interface{} is accepted here.
func ParseService(ifacePtr interface{}) (*ServiceInfo, error) {
	t := reflect.TypeOf(ifacePtr)
	if t == nil {
		return nil, fmt.Errorf("input is nil")
	}

	// If (*InterfaceType)(nil) is passed, Elem() gets InterfaceType
	if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Interface {
		t = t.Elem()
	} else if t.Kind() != reflect.Interface {
		return nil, fmt.Errorf("input must be a pointer to an interface type (e.g., (*MyInterface)(nil))")
	}

	pkgPath := t.PkgPath()
	typeName := t.Name()
	if typeName == "" {
		typeName = "<anonymous>"
	}

	var methods []MethodInfo
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		method := MethodInfo{
			Name:   m.Name,
			Input:  parseType(m.Type.In(1)),
			Output: parseType(m.Type.Out(0)),
		}
		methods = append(methods, method)
	}

	return &ServiceInfo{
		Name:    typeName,
		Package: pkgPath,
		Methods: methods,
	}, nil
}

// parseType parses a reflect.Type to get detailed type information.
func parseType(t reflect.Type) *TypeInfo {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	info := &TypeInfo{
		Name: t.Name(),
	}

	if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			// Skip unexported fields
			if f.PkgPath != "" {
				continue
			}
			info.Fields = append(info.Fields, FieldInfo{
				Name: f.Name,
				Type: f.Type.String(),
				Tag:  string(f.Tag),
			})
		}
	}
	return info
}

// String implements the string representation of ServiceInfo (for easy printing).
func (ii *ServiceInfo) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("interface %s {\n", ii.Name))
	for _, m := range ii.Methods {
		sb.WriteString(fmt.Sprintf("    %s(%s) %s\n", m.Name, m.Input.Name, m.Output.Name))
	}
	sb.WriteString("}")
	return sb.String()
}
