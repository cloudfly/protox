package doc

import (
	"fmt"
	"reflect"
	"strings"
)

// ServiceInfo 表示接口的信息
type ServiceInfo struct {
	Name    string
	Package string
	Methods []MethodInfo
}

// MethodInfo 表示接口中一个方法的定义
type MethodInfo struct {
	Name   string
	Input  reflect.Type // 输入参数类型名
	Output reflect.Type // 返回值类型名
}

// ParseService 解析 interface 类型的方法信息
// 注意：input 必须是一个指向 interface 类型的指针（如 (*MyInterface)(nil)）
// 或者传入 reflect.Type（更推荐），但为简化 API，这里接受 interface{}
func ParseService(ifacePtr interface{}) (*ServiceInfo, error) {
	t := reflect.TypeOf(ifacePtr)
	if t == nil {
		return nil, fmt.Errorf("input is nil")
	}

	// 如果传入的是 (*InterfaceType)(nil)，则 Elem() 得到 InterfaceType
	if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Interface {
		t = t.Elem()
	} else if t.Kind() != reflect.Interface {
		return nil, fmt.Errorf("input must be a pointer to an interface type (e.g., (*MyInterface)(nil))")
	}

	// 获取包路径和类型名
	pkgPath := t.PkgPath()
	typeName := t.Name()
	if typeName == "" {
		// 匿名接口（如 interface{ Foo() }）没有名字
		typeName = "<anonymous>"
	}

	var methods []MethodInfo
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		method := MethodInfo{
			Name:   m.Name,
			Input:  m.Type.In(1).Elem(),
			Output: m.Type.Out(0).Elem(),
		}
		methods = append(methods, method)
	}

	return &ServiceInfo{
		Name:    typeName,
		Package: pkgPath,
		Methods: methods,
	}, nil
}

// getTypeName 返回类型的可读字符串表示
func getTypeName(t reflect.Type) string {
	// 处理指针
	if t.Kind() == reflect.Ptr {
		return "*" + getTypeName(t.Elem())
	}

	// 处理切片
	if t.Kind() == reflect.Slice {
		return "[]" + getTypeName(t.Elem())
	}

	// 处理 map
	if t.Kind() == reflect.Map {
		return fmt.Sprintf("map[%s]%s", getTypeName(t.Key()), getTypeName(t.Elem()))
	}

	// 处理通道
	if t.Kind() == reflect.Chan {
		dir := ""
		switch t.ChanDir() {
		case reflect.SendDir:
			dir = "chan<- "
		case reflect.RecvDir:
			dir = "<-chan "
		default:
			dir = "chan "
		}
		return dir + getTypeName(t.Elem())
	}

	// 如果有包路径且不是内置类型
	if t.PkgPath() != "" && t.Name() != "" {
		return t.PkgPath() + "." + t.Name()
	}

	// 内置类型（int, string, bool 等）
	return t.Kind().String()
}

// String 实现 InterfaceInfo 的字符串表示（便于打印）
func (ii *ServiceInfo) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("interface %s {\n", ii.Name))
	for _, m := range ii.Methods {
		sb.WriteString(fmt.Sprintf("    %s(%s) %s\n", m.Name, m.Input.Name(), m.Output.Name()))
	}
	sb.WriteString("}")
	return sb.String()
}
