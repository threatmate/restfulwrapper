package restfulwrapper

import (
	"fmt"
	"reflect"
	"strconv"
)

// ParameterParser is an interface that a parameter can implement in order to
// be eligible for use in query and path parameters.
type ParameterParser interface {
	// Parse accepts a string and returns an error on failure.
	ParseString(input string) error
}

// parseStringToSingleValue parses a string value into the target given.
//
// This will return an error if `target` is not a pointer or if it is nil.
//
// This supports all of the Go primitives, such as int, uint64, string, etc.
func parseStringToSingleValue(stringValue string, target any) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Pointer || targetValue.IsNil() {
		return fmt.Errorf("invalid target: needed pointer, got %s", targetValue.Kind().String())
	}

	if targetValue.CanInterface() {
		parser, ok := targetValue.Interface().(ParameterParser)
		if ok {
			err := parser.ParseString(stringValue)
			if err != nil {
				return fmt.Errorf("could not parse string value: %w", err)
			}
			return nil
		}
	}

	switch targetValue.Elem().Kind() {
	case reflect.Bool:
		v, err := strconv.ParseBool(stringValue)
		if err != nil {
			return err
		}
		targetValue.Elem().SetBool(v)
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(stringValue, targetValue.Elem().Type().Bits())
		if err != nil {
			return err
		}
		targetValue.Elem().SetFloat(v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(stringValue, 10, targetValue.Elem().Type().Bits())
		if err != nil {
			return err
		}
		targetValue.Elem().SetInt(v)
	case reflect.String:
		targetValue.Elem().SetString(stringValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(stringValue, 10, targetValue.Elem().Type().Bits())
		if err != nil {
			return err
		}
		targetValue.Elem().SetUint(v)
	default:
		return fmt.Errorf("could not parse to single value: unhandled kind: %s", targetValue.Elem().Kind().String())
	}

	return nil
}
