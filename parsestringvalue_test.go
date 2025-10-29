package restfulwrapper

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseStringValue(t *testing.T) {
	type MyFloat64 float64

	rows := []struct {
		Description string
		Input       string
		Target      any
		Success     bool
		Output      any
	}{
		{
			Description: "target string must be a pointer",
			Input:       "",
			Target:      "",
			Success:     false,
		},
		{
			Description: "target int must be a pointer",
			Input:       "",
			Target:      int(0),
			Success:     false,
		},
		{
			Description: "target float64 must be a pointer",
			Input:       "",
			Target:      float64(0),
			Success:     false,
		},
		{
			Description: "string can be empty",
			Input:       "",
			Target:      new(string),
			Success:     true,
			Output:      "",
		},
		{
			Description: "string can have spaces",
			Input:       "hello, world",
			Target:      new(string),
			Success:     true,
			Output:      "hello, world",
		},
		{
			Description: "bool cannot be empty",
			Input:       "",
			Target:      new(bool),
			Success:     false,
		},
		{
			Description: "bool can be true",
			Input:       "true",
			Target:      new(bool),
			Success:     true,
			Output:      true,
		},
		{
			Description: "int cannot be empty",
			Input:       "",
			Target:      new(int),
			Success:     false,
		},
		{
			Description: "int cannot be words",
			Input:       "hello, world",
			Target:      new(int),
			Success:     false,
		},
		{
			Description: "int cannot start with words",
			Input:       "hello 9",
			Target:      new(int),
			Success:     false,
		},
		{
			Description: "int cannot start with words",
			Input:       "9 world",
			Target:      new(int),
			Success:     false,
		},
		{
			Description: "int can be 0",
			Input:       "0",
			Target:      new(int),
			Success:     true,
			Output:      int(0),
		},
		{
			Description: "int can be 1234",
			Input:       "1234",
			Target:      new(int),
			Success:     true,
			Output:      int(1234),
		},
		{
			Description: "int cannot have fractional parts",
			Input:       "1234.5",
			Target:      new(int),
			Success:     false,
		},
		{
			Description: "float64 cannot be empty",
			Input:       "",
			Target:      new(float64),
			Success:     false,
		},
		{
			Description: "float64 can be 0",
			Input:       "0",
			Target:      new(float64),
			Success:     true,
			Output:      float64(0),
		},
		{
			Description: "float64 can be 1234",
			Input:       "1234",
			Target:      new(float64),
			Success:     true,
			Output:      float64(1234),
		},
		{
			Description: "float64 can have fractional parts",
			Input:       "1234.5",
			Target:      new(float64),
			Success:     true,
			Output:      float64(1234.5),
		},
		{
			Description: "uint can be 1234",
			Input:       "1234",
			Target:      new(uint),
			Success:     true,
			Output:      uint(1234),
		},
		{
			Description: "uint8 can be 123",
			Input:       "123",
			Target:      new(uint8),
			Success:     true,
			Output:      uint8(123),
		},
		{
			Description: "uint8 cannot be 1234",
			Input:       "1234",
			Target:      new(uint8),
			Success:     false,
		},
		{
			Description: "uint16 can be 1234",
			Input:       "1234",
			Target:      new(uint16),
			Success:     true,
			Output:      uint16(1234),
		},
		{
			Description: "uint32 can be 1234",
			Input:       "1234",
			Target:      new(uint32),
			Success:     true,
			Output:      uint32(1234),
		},
		{
			Description: "uint64 can be 1234",
			Input:       "1234",
			Target:      new(uint64),
			Success:     true,
			Output:      uint64(1234),
		},
		{
			Description: "MyFloat64 can have fractional parts",
			Input:       "1234.5",
			Target:      new(MyFloat64),
			Success:     true,
			Output:      MyFloat64(1234.5),
		},
		{
			Description: "target cannot be a struct",
			Input:       "",
			Target:      new(struct{}),
			Success:     false,
		},
	}
	for rowIndex, row := range rows {
		t.Run(fmt.Sprintf("%d/%s", rowIndex, row.Description), func(t *testing.T) {
			err := parseStringToSingleValue(row.Input, row.Target)
			if !row.Success {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			assert.Equal(t, row.Output, reflect.ValueOf(row.Target).Elem().Interface())
		})
	}
}
