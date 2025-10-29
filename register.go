package restfulwrapper

import (
	"fmt"
	"reflect"
)

// RegisterFunction is a function that can be used to register a new API tag.
type RegisterFunction func(apiTagValue string, field reflect.StructField, info *RestfulFunctionInfo) (InputFieldFunction, error)

// registeredFunctionMap is the map of registered API tags to their functions.
var registeredFunctionMap = map[string]RegisterFunction{}

// Register a new API tag.
//
// This will panic if the tag is already registered.
func Register(apiTagKey string, f RegisterFunction) {
	if _, ok := registeredFunctionMap[apiTagKey]; ok {
		panic(fmt.Errorf("tag already registered: %s", apiTagKey))
	}
	registeredFunctionMap[apiTagKey] = f
}
