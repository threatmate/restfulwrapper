package restfulwrapper

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

// ParseRestfulFunction accepts a function and returns the parsed information about that function.
//
// This information can be used to generate a function that can be called to handle the given REST
// request.
func ParseRestfulFunction(f interface{}) (*RestfulFunctionInfo, error) {
	if f == nil {
		return nil, fmt.Errorf("expected function; got nil")
	}
	if reflect.TypeOf(f).Kind() != reflect.Func {
		return nil, fmt.Errorf("expected function; got %v", reflect.TypeOf(f).Kind())
	}

	info := RestfulFunctionInfo{
		FunctionValue:       reflect.ValueOf(f),
		InContextPosition:   -1,
		InMetadataPosition:  -1,
		OutErrorPosition:    -1,
		OutResponsePosition: -1,
		LocalMap:            map[string]string{},
	}

	for i := range info.FunctionValue.Type().NumIn() {
		argumentType := info.FunctionValue.Type().In(i)

		contextType := reflect.TypeOf((*context.Context)(nil)).Elem()

		if argumentType.Implements(contextType) {
			if info.InContextPosition >= 0 {
				return nil, fmt.Errorf("multiple context.Context arguments")
			}
			info.InContextPosition = i
		} else {
			if info.InMetadataPosition >= 0 {
				return nil, fmt.Errorf("multiple input arguments")
			}
			info.InMetadataPosition = i
			info.InMetadataType = argumentType
		}
	}

	for i := range info.FunctionValue.Type().NumOut() {
		argumentType := info.FunctionValue.Type().Out(i)

		errorType := reflect.TypeOf((*error)(nil)).Elem()

		if argumentType.Implements(errorType) {
			if info.OutErrorPosition >= 0 {
				return nil, fmt.Errorf("multiple error arguments")
			}
			info.OutErrorPosition = i
		} else {
			if info.OutResponsePosition >= 0 {
				return nil, fmt.Errorf("multiple output arguments")
			}
			info.OutResponsePosition = i
		}
	}

	if info.OutResponsePosition >= 0 {
		argumentType := info.FunctionValue.Type().Out(info.OutResponsePosition)

		exampleValue := reflect.New(argumentType)
		if exampleValue.CanAddr() { // TODO: Is this necessary?
			exampleValue = exampleValue.Addr()
		}
		info.ResponseExample = exampleValue.Interface()
	}

	if info.InMetadataPosition >= 0 {
		argumentType := info.FunctionValue.Type().In(info.InMetadataPosition)
		switch argumentType.Kind() {
		case reflect.Pointer:
			argumentType = argumentType.Elem()
		}

		if argumentType.Kind() != reflect.Struct {
			return nil, fmt.Errorf("unexpected input type: %v", argumentType.Kind())
		}

		for fieldIndex := range argumentType.NumField() {
			field := argumentType.Field(fieldIndex)
			err := handleField(&info, field)
			if err != nil {
				return nil, fmt.Errorf("could not handle field %q: %w", field.Name, err)
			}
		}
	}

	return &info, nil
}

func handleField(info *RestfulFunctionInfo, field reflect.StructField) error {
	// "Anonymous" fields are when you embed a struct.
	//
	// When we have an anonymous field, go through all of *its* fields and add them.
	if field.Anonymous {
		for i := range field.Type.NumField() {
			err := handleField(info, field.Type.Field(i))
			if err != nil {
				return err
			}
		}
		return nil
	}

	apiTagText := field.Tag.Get("api")
	if apiTagText == "" {
		return nil
	}

	parts := strings.SplitN(apiTagText, ":", 2)
	apiTagKey := parts[0]
	apiTagValue := ""
	if len(parts) > 1 {
		apiTagValue = parts[1]
	}

	inputField := InputField{
		Name: field.Name,
	}

	if apiTagKey == "-" {
		return nil
	}

	registeredFunction := registeredFunctionMap[apiTagKey]
	if registeredFunction == nil {
		return fmt.Errorf("unhandled API tag: %s", apiTagKey)
	}

	inputFieldFunction, err := registeredFunction(apiTagValue, field, info)
	if err != nil {
		return fmt.Errorf("%s: %w", field.Name, err)
	}
	inputField.Function = inputFieldFunction

	if inputField.Function != nil {
		info.InputFields = append(info.InputFields, inputField)
	}

	return nil
}
