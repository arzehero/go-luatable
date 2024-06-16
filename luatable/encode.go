package luatable

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
)

var timeType = reflect.TypeOf(time.Time{})
var EmptyTable = "return {}"

func Encode(values interface{}) (string, error) {
	value := reflect.ValueOf(values)
	for value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return EmptyTable, nil
		}
		value = value.Elem()
	}

	if values == nil {
		return EmptyTable, nil
	}

	if value.Kind() == reflect.Array || value.Kind() == reflect.Slice {
		result, err := reflectStringifyArray(value, 1)
		if err != nil {
			return EmptyTable, err
		}

		encoded := fmt.Sprintf("return %s", result)
		return encoded, nil
	}

	if value.Kind() == reflect.Map {
		result, err := reflectStringifyMap(value, 1)
		if err != nil {
			return EmptyTable, err
		}

		encoded := fmt.Sprintf("return %s", result)
		return encoded, nil
	}

	if value.Kind() != reflect.Struct {
		return EmptyTable, fmt.Errorf("luatable: Encode() expects struct input. Got %v", value.Kind())
	}

	result := new(bytes.Buffer)
	result, err := reflectStringifyStruct(value, 1)
	if err != nil {
		return EmptyTable, err
	}

	encoded := fmt.Sprintf("return %s", result.String())
	return encoded, nil
}

func reflectStringifyArray(value reflect.Value, indentSize int) (string, error) {
	if value.Len() == 0 {
		return "{}", nil
	}

	result := new(bytes.Buffer)
	result.WriteString("{")

	first := value.Index(0)
	newline := " "
	indent := strings.Repeat(" ", 2*indentSize)

	if first.Kind() == reflect.Struct ||
		first.Kind() == reflect.Array ||
		first.Kind() == reflect.Slice ||
		first.Kind() == reflect.String {
		newline = "\n"
	} else {
		indent = ""
	}

	result.WriteString(newline)
	for i := 0; i < value.Len(); i++ {
		item := value.Index(i)

		var stringified string
		var err error
		switch item.Kind() {
		case reflect.Array, reflect.Slice:
			stringified, err = reflectStringifyArray(item, indentSize+1)
			if err != nil {
				return "{}", err
			}
		case reflect.Struct:
			buffer, err := reflectStringifyStruct(item, indentSize+1)
			if err != nil {
				return "{}", err
			}

			stringified = buffer.String()
		default:
			stringified = valueString(item, tagOptions{}, reflect.StructField{})
		}

		val := fmt.Sprintf("%s%s,%s", indent, stringified, newline)
		result.WriteString(val)
	}

	if newline != " " {
		indent := strings.Repeat(" ", 2*(indentSize-1))
		result.WriteString(indent)
	}
	result.WriteString("}")
	return result.String(), nil
}

func reflectStringifyMap(value reflect.Value, indentSize int) (string, error) {
	if value.Len() == 0 {
		return "{}", nil
	}

	result := new(bytes.Buffer)
	result.WriteString("{\n")
	iter := value.MapRange()

	for iter.Next() {
		key := iter.Key()
		value := iter.Value()

		var strValue string
		switch value.Kind() {
		case reflect.Struct:
			buffer, err := reflectStringifyStruct(value, indentSize+1)
			if err != nil {
				return result.String(), err
			}
			strValue = buffer.String()
		case reflect.Array, reflect.Slice:
			var err error
			strValue, err = reflectStringifyArray(value, indentSize+1)
			if err != nil {
				return result.String(), err
			}
		case reflect.Map:
			var err error
			strValue, err = reflectStringifyMap(value, indentSize+1)
			if err != nil {
				return result.String(), err
			}
		default:
			strValue = valueString(value, tagOptions{}, reflect.StructField{})
		}

		writeValue(result, key.String(), strValue, indentSize)
	}

	result.WriteString(strings.Repeat(" ", 2*(indentSize-1)) + "}")
	return result.String(), nil
}

func reflectStringifyStruct(value reflect.Value, indentSize int) (*bytes.Buffer, error) {
	result := new(bytes.Buffer)
	result.WriteString("{\n")
	structType := value.Type()

	for i := 0; i < structType.NumField(); i++ {
		structField := structType.Field(i)
		if structField.PkgPath != "" && !structField.Anonymous {
			continue
		}

		structValue := value.Field(i)
		tag := structField.Tag.Get("lua")
		if tag == "-" {
			continue
		}
		tagName, tagOptions := parseTag(tag)

		if tagOptions.Contains("omitempty") && isEmptyValue(structValue) {
			continue
		}

		for structValue.Kind() == reflect.Ptr {
			if structValue.IsNil() {
				break
			}
			structValue = structValue.Elem()
		}

		var strValue string
		switch structValue.Kind() {
		case reflect.Struct:
			buffer, err := reflectStringifyStruct(structValue, indentSize+1)
			if err != nil {
				return result, err
			}
			strValue = buffer.String()
		case reflect.Array, reflect.Slice:
			var err error
			strValue, err = reflectStringifyArray(structValue, indentSize+1)
			if err != nil {
				return result, err
			}
		case reflect.Map:
			var err error
			strValue, err = reflectStringifyMap(structValue, indentSize+1)
			if err != nil {
				return result, err
			}
		default:
			strValue = valueString(structValue, tagOptions, structField)
		}

		writeValue(result, tagName, strValue, indentSize)
	}

	indent := strings.Repeat(" ", 2*(indentSize-1))
	result.WriteString(indent + "}")
	return result, nil
}

func valueString(value reflect.Value, options tagOptions, field reflect.StructField) string {
	for value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return "nil"
		}
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.String:
		str := strings.ReplaceAll(value.String(), "'", "\\'")
		str = strings.ReplaceAll(str, "\n", "\\n")
		return fmt.Sprintf("'%s'", str)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprint(value.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprint(value.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprint(value.Float())
	case reflect.Bool:
		return fmt.Sprint(value.Bool())
	}

	return "nil"
}

var isValidKeyName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`).MatchString

func writeValue(result *bytes.Buffer, name string, value string, indentSize int) {
	indent := strings.Repeat(" ", 2*indentSize)
	result.WriteString(indent)

	if isValidKeyName(name) {
		result.WriteString(name)
	} else {
		escaped := strings.ReplaceAll(name, "'", "\\'")
		key := fmt.Sprintf("['%s']", escaped)
		result.WriteString(key)
	}

	result.WriteString(fmt.Sprintf(" = %s,\n", value))
}

type tagOptions []string

func (options tagOptions) Contains(option string) bool {
	for _, value := range options {
		if value == option {
			return true
		}
	}
	return false
}

func parseTag(tag string) (string, tagOptions) {
	s := strings.Split(tag, ",")
	return s[0], s[1:]
}

func isEmptyValue(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return value.Len() == 0
	case reflect.Bool:
		return !value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return value.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return value.IsNil()
	}

	return false
}
