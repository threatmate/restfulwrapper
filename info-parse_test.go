package restfulwrapper

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRestfulFunction(t *testing.T) {
	t.Run("Not a function", func(t *testing.T) {
		rows := []any{
			nil,
			0,
			1,
			"a",
			"hello",
			struct{}{},
		}
		for rowIndex, row := range rows {
			t.Run(fmt.Sprintf("%d", rowIndex), func(t *testing.T) {
				output, err := ParseRestfulFunction(row)
				require.NotNil(t, err)
				assert.Nil(t, output)
			})
		}
	})
	t.Run("Function", func(t *testing.T) {
		t.Run("Trivial", func(t *testing.T) {
			input := func() {}
			output, err := ParseRestfulFunction(input)
			require.Nil(t, err)
			assert.NotNil(t, output)
			assert.Nil(t, output.BodyExample)
			assert.Equal(t, -1, output.InContextPosition)
			assert.Equal(t, -1, output.InMetadataPosition)
			assert.Nil(t, output.InMetadataType)
			assert.Equal(t, -1, output.OutErrorPosition)
			assert.Equal(t, -1, output.OutResponsePosition)
			assert.Nil(t, output.ResponseExample)

			assert.Equal(t, 0, len(output.InputFields))
			assert.Equal(t, 0, len(output.PathParameters))
			assert.Equal(t, 0, len(output.QueryParameters))
		})
		t.Run("With metadata", func(t *testing.T) {
			input := func(struct{}) {}
			output, err := ParseRestfulFunction(input)
			require.Nil(t, err)
			assert.NotNil(t, output)
			assert.Nil(t, output.BodyExample)
			assert.Equal(t, -1, output.InContextPosition)
			assert.Equal(t, 0, output.InMetadataPosition)
			assert.NotNil(t, output.InMetadataType)
			assert.Equal(t, -1, output.OutErrorPosition)
			assert.Equal(t, -1, output.OutResponsePosition)
			assert.Nil(t, output.ResponseExample)

			assert.Equal(t, 0, len(output.InputFields))
			assert.Equal(t, 0, len(output.PathParameters))
			assert.Equal(t, 0, len(output.QueryParameters))
		})
		t.Run("With metadata returns error", func(t *testing.T) {
			input := func(struct{}) error { return nil }
			output, err := ParseRestfulFunction(input)
			require.Nil(t, err)
			assert.NotNil(t, output)
			assert.Nil(t, output.BodyExample)
			assert.Equal(t, -1, output.InContextPosition)
			assert.Equal(t, 0, output.InMetadataPosition)
			assert.NotNil(t, output.InMetadataType)
			assert.Equal(t, 0, output.OutErrorPosition)
			assert.Equal(t, -1, output.OutResponsePosition)
			assert.Nil(t, output.ResponseExample)

			assert.Equal(t, 0, len(output.InputFields))
			assert.Equal(t, 0, len(output.PathParameters))
			assert.Equal(t, 0, len(output.QueryParameters))
		})
		t.Run("With context and metadata returns response and error", func(t *testing.T) {
			input := func(context.Context, struct{}) (string, error) { return "", nil }
			output, err := ParseRestfulFunction(input)
			require.Nil(t, err)
			assert.NotNil(t, output)
			assert.Nil(t, output.BodyExample)
			assert.Equal(t, 0, output.InContextPosition)
			assert.Equal(t, 1, output.InMetadataPosition)
			assert.NotNil(t, output.InMetadataType)
			assert.Equal(t, 1, output.OutErrorPosition)
			assert.Equal(t, 0, output.OutResponsePosition)
			assert.NotNil(t, output.ResponseExample)

			assert.Equal(t, 0, len(output.InputFields))
			assert.Equal(t, 0, len(output.PathParameters))
			assert.Equal(t, 0, len(output.QueryParameters))
		})
		t.Run("With context and metadata returns pointer response and error", func(t *testing.T) {
			input := func(context.Context, struct{}) (*string, error) { return nil, nil }
			output, err := ParseRestfulFunction(input)
			require.Nil(t, err)
			assert.NotNil(t, output)
			assert.Nil(t, output.BodyExample)
			assert.Equal(t, 0, output.InContextPosition)
			assert.Equal(t, 1, output.InMetadataPosition)
			assert.NotNil(t, output.InMetadataType)
			assert.Equal(t, 1, output.OutErrorPosition)
			assert.Equal(t, 0, output.OutResponsePosition)
			assert.NotNil(t, output.ResponseExample)

			assert.Equal(t, 0, len(output.InputFields))
			assert.Equal(t, 0, len(output.PathParameters))
			assert.Equal(t, 0, len(output.QueryParameters))
		})
		t.Run("With context and pointer metadata returns pointer response and error", func(t *testing.T) {
			input := func(context.Context, *struct{}) (*string, error) { return nil, nil }
			output, err := ParseRestfulFunction(input)
			require.Nil(t, err)
			assert.NotNil(t, output)
			assert.Nil(t, output.BodyExample)
			assert.Equal(t, 0, output.InContextPosition)
			assert.Equal(t, 1, output.InMetadataPosition)
			assert.NotNil(t, output.InMetadataType)
			assert.Equal(t, 1, output.OutErrorPosition)
			assert.Equal(t, 0, output.OutResponsePosition)
			assert.NotNil(t, output.ResponseExample)

			assert.Equal(t, 0, len(output.InputFields))
			assert.Equal(t, 0, len(output.PathParameters))
			assert.Equal(t, 0, len(output.QueryParameters))
		})
		t.Run("Too many contexts", func(t *testing.T) {
			input := func(context.Context, context.Context) {}
			output, err := ParseRestfulFunction(input)
			require.NotNil(t, err)
			assert.Nil(t, output)
		})
		t.Run("Too many metadatas", func(t *testing.T) {
			input := func(struct{}, struct{}) {}
			output, err := ParseRestfulFunction(input)
			require.NotNil(t, err)
			assert.Nil(t, output)
		})
		t.Run("Incorrect metadata", func(t *testing.T) {
			input := func(string) {}
			output, err := ParseRestfulFunction(input)
			require.NotNil(t, err)
			assert.Nil(t, output)
		})
		t.Run("Too many responses", func(t *testing.T) {
			input := func() (string, string) { return "", "" }
			output, err := ParseRestfulFunction(input)
			require.NotNil(t, err)
			assert.Nil(t, output)
		})
		t.Run("Too many errors", func(t *testing.T) {
			input := func() (error, error) { return nil, nil }
			output, err := ParseRestfulFunction(input)
			require.NotNil(t, err)
			assert.Nil(t, output)
		})
		t.Run("Unhandled fields", func(t *testing.T) {
			input := func(struct {
				NotUsedInt    int
				NotUsedString string
			}) {
			}
			output, err := ParseRestfulFunction(input)
			require.Nil(t, err)
			assert.NotNil(t, output)
			assert.Nil(t, output.BodyExample)
			assert.Equal(t, -1, output.InContextPosition)
			assert.Equal(t, 0, output.InMetadataPosition)
			assert.NotNil(t, output.InMetadataType)
			assert.Equal(t, -1, output.OutErrorPosition)
			assert.Equal(t, -1, output.OutResponsePosition)
			assert.Nil(t, output.ResponseExample)

			assert.Equal(t, 0, len(output.InputFields))
			assert.Equal(t, 0, len(output.PathParameters))
			assert.Equal(t, 0, len(output.QueryParameters))
		})
		t.Run("Unhandled fields explicit", func(t *testing.T) {
			input := func(struct {
				NotUsedInt    int    `api:"-"`
				NotUsedString string `api:"-"`
			}) {
			}
			output, err := ParseRestfulFunction(input)
			require.Nil(t, err)
			assert.NotNil(t, output)
			assert.Nil(t, output.BodyExample)
			assert.Equal(t, -1, output.InContextPosition)
			assert.Equal(t, 0, output.InMetadataPosition)
			assert.NotNil(t, output.InMetadataType)
			assert.Equal(t, -1, output.OutErrorPosition)
			assert.Equal(t, -1, output.OutResponsePosition)
			assert.Nil(t, output.ResponseExample)

			assert.Equal(t, 0, len(output.InputFields))
			assert.Equal(t, 0, len(output.PathParameters))
			assert.Equal(t, 0, len(output.QueryParameters))
		})
		t.Run("Bogus field", func(t *testing.T) {
			input := func(struct {
				Bogus string `api:"bogus"`
			}) {
			}
			output, err := ParseRestfulFunction(input)
			require.NotNil(t, err)
			assert.Nil(t, output)
		})
		t.Run("body", func(t *testing.T) {
			t.Run("string", func(t *testing.T) {
				input := func(struct {
					Body string `api:"body"`
				}) {
				}
				output, err := ParseRestfulFunction(input)
				require.Nil(t, err)
				assert.NotNil(t, output)
				assert.NotNil(t, output.BodyExample)
				assert.Equal(t, -1, output.InContextPosition)
				assert.Equal(t, 0, output.InMetadataPosition)
				assert.NotNil(t, output.InMetadataType)
				assert.Equal(t, -1, output.OutErrorPosition)
				assert.Equal(t, -1, output.OutResponsePosition)
				assert.Nil(t, output.ResponseExample)

				assert.Equal(t, 1, len(output.InputFields))
				assert.Equal(t, 0, len(output.PathParameters))
				assert.Equal(t, 0, len(output.QueryParameters))
			})
			t.Run("bad syntax", func(t *testing.T) {
				input := func(struct {
					Body string `api:"body:extra"`
				}) {
				}
				output, err := ParseRestfulFunction(input)
				require.NotNil(t, err)
				assert.Nil(t, output)
			})
			t.Run("good multipart form", func(t *testing.T) {
				input := func(struct {
					Body multipart.Form `api:"body"`
				}) {
				}
				output, err := ParseRestfulFunction(input)
				require.Nil(t, err)
				require.NotNil(t, output)
				assert.Equal(t, []string{"multipart/form-data"}, output.Consumes)
			})
			t.Run("good multipart form pointer", func(t *testing.T) {
				input := func(struct {
					Body *multipart.Form `api:"body"`
				}) {
				}
				output, err := ParseRestfulFunction(input)
				require.Nil(t, err)
				require.NotNil(t, output)
				assert.Equal(t, []string{"multipart/form-data"}, output.Consumes)
			})
			t.Run("bad multipart form", func(t *testing.T) {
				input := func(struct {
					Body string `api:"body:consumes:multipart/form-data"`
				}) {
				}
				output, err := ParseRestfulFunction(input)
				require.NotNil(t, err)
				assert.Nil(t, output)
			})
			t.Run("good values form", func(t *testing.T) {
				input := func(struct {
					Body url.Values `api:"body"`
				}) {
				}
				output, err := ParseRestfulFunction(input)
				require.Nil(t, err)
				require.NotNil(t, output)
				assert.Equal(t, []string{"application/x-www-form-urlencoded"}, output.Consumes)
			})
			t.Run("good values form pointer", func(t *testing.T) {
				input := func(struct {
					Body *url.Values `api:"body"`
				}) {
				}
				output, err := ParseRestfulFunction(input)
				require.Nil(t, err)
				require.NotNil(t, output)
				assert.Equal(t, []string{"application/x-www-form-urlencoded"}, output.Consumes)
			})
			t.Run("bad values form", func(t *testing.T) {
				input := func(struct {
					Body string `api:"body:consumes:application/x-www-form-urlencoded"`
				}) {
				}
				output, err := ParseRestfulFunction(input)
				require.NotNil(t, err)
				assert.Nil(t, output)
			})
		})
		t.Run("httprequest", func(t *testing.T) {
			t.Run("good httprequest", func(t *testing.T) {
				input := func(struct {
					Request *http.Request `api:"httprequest"`
				}) {
				}
				output, err := ParseRestfulFunction(input)
				require.Nil(t, err)
				assert.NotNil(t, output)
				assert.Nil(t, output.BodyExample)
				assert.Equal(t, -1, output.InContextPosition)
				assert.Equal(t, 0, output.InMetadataPosition)
				assert.NotNil(t, output.InMetadataType)
				assert.Equal(t, -1, output.OutErrorPosition)
				assert.Equal(t, -1, output.OutResponsePosition)
				assert.Nil(t, output.ResponseExample)

				assert.Equal(t, 1, len(output.InputFields))
				assert.Equal(t, 0, len(output.PathParameters))
				assert.Equal(t, 0, len(output.QueryParameters))
			})
			t.Run("Bad httprequest", func(t *testing.T) {
				input := func(struct {
					Request *http.Request `api:"httprequest:extra"`
				}) {
				}
				output, err := ParseRestfulFunction(input)
				require.NotNil(t, err)
				assert.Nil(t, output)
			})
			t.Run("Bad httprequest type", func(t *testing.T) {
				input := func(struct {
					Request string `api:"httprequest"`
				}) {
				}
				output, err := ParseRestfulFunction(input)
				require.NotNil(t, err)
				assert.Nil(t, output)
			})
		})
		t.Run("path", func(t *testing.T) {
			t.Run("good path", func(t *testing.T) {
				input := func(struct {
					Value1 string `api:"path:key1" description:"my description"`
				}) {
				}
				output, err := ParseRestfulFunction(input)
				require.Nil(t, err)
				assert.NotNil(t, output)
				assert.Nil(t, output.BodyExample)
				assert.Equal(t, -1, output.InContextPosition)
				assert.Equal(t, 0, output.InMetadataPosition)
				assert.NotNil(t, output.InMetadataType)
				assert.Equal(t, -1, output.OutErrorPosition)
				assert.Equal(t, -1, output.OutResponsePosition)
				assert.Nil(t, output.ResponseExample)

				assert.Equal(t, 1, len(output.InputFields))
				if assert.Equal(t, 1, len(output.PathParameters)) {
					assert.Equal(t, "Value1", output.PathParameters[0].FieldName)
					assert.Equal(t, "key1", output.PathParameters[0].Name)
					assert.Equal(t, "my description", output.PathParameters[0].Description)
				}
				assert.Equal(t, 0, len(output.QueryParameters))
			})
			t.Run("Bad path", func(t *testing.T) {
				input := func(struct {
					Value1 string `api:"path"`
				}) {
				}
				output, err := ParseRestfulFunction(input)
				require.NotNil(t, err)
				assert.Nil(t, output)
			})
			t.Run("Duplicate path", func(t *testing.T) {
				input := func(struct {
					Value1 string `api:"path:key1"`
					Value2 string `api:"path:key1"`
				}) {
				}
				output, err := ParseRestfulFunction(input)
				require.NotNil(t, err)
				assert.Nil(t, output)
			})
		})
		t.Run("query", func(t *testing.T) {
			t.Run("good query", func(t *testing.T) {
				input := func(struct {
					Value1 string `api:"query:key1" description:"my description"`
				}) {
				}
				output, err := ParseRestfulFunction(input)
				require.Nil(t, err)
				assert.NotNil(t, output)
				assert.Nil(t, output.BodyExample)
				assert.Equal(t, -1, output.InContextPosition)
				assert.Equal(t, 0, output.InMetadataPosition)
				assert.NotNil(t, output.InMetadataType)
				assert.Equal(t, -1, output.OutErrorPosition)
				assert.Equal(t, -1, output.OutResponsePosition)
				assert.Nil(t, output.ResponseExample)

				assert.Equal(t, 1, len(output.InputFields))
				assert.Equal(t, 0, len(output.PathParameters))
				if assert.Equal(t, 1, len(output.QueryParameters)) {
					assert.Equal(t, "Value1", output.QueryParameters[0].FieldName)
					assert.Equal(t, "key1", output.QueryParameters[0].Name)
					assert.Equal(t, "my description", output.QueryParameters[0].Description)
				}
			})
			t.Run("deprecated query", func(t *testing.T) {
				input := func(struct {
					Value1 string `api:"query:key1,oldKey" description:"my description"`
				}) {
				}
				output, err := ParseRestfulFunction(input)
				require.Nil(t, err)
				assert.NotNil(t, output)
				assert.Nil(t, output.BodyExample)
				assert.Equal(t, -1, output.InContextPosition)
				assert.Equal(t, 0, output.InMetadataPosition)
				assert.NotNil(t, output.InMetadataType)
				assert.Equal(t, -1, output.OutErrorPosition)
				assert.Equal(t, -1, output.OutResponsePosition)
				assert.Nil(t, output.ResponseExample)

				assert.Equal(t, 1, len(output.InputFields))
				assert.Equal(t, 0, len(output.PathParameters))
				if assert.Equal(t, 2, len(output.QueryParameters)) {
					assert.Equal(t, "Value1", output.QueryParameters[0].FieldName)
					assert.Equal(t, "key1", output.QueryParameters[0].Name)
					assert.Equal(t, "my description", output.QueryParameters[0].Description)

					assert.Equal(t, "Value1", output.QueryParameters[1].FieldName)
					assert.Equal(t, "oldKey", output.QueryParameters[1].Name)
					assert.Equal(t, `Deprecated; use "key1" instead.`, output.QueryParameters[1].Description)
				}
			})
			t.Run("Bad query", func(t *testing.T) {
				input := func(struct {
					Value1 string `api:"query"`
				}) {
				}
				output, err := ParseRestfulFunction(input)
				require.NotNil(t, err)
				assert.Nil(t, output)
			})
			t.Run("Duplicate query", func(t *testing.T) {
				input := func(struct {
					Value1 string `api:"query:key1"`
					Value2 string `api:"query:key1"`
				}) {
				}
				output, err := ParseRestfulFunction(input)
				require.NotNil(t, err)
				assert.Nil(t, output)
			})
		})
		t.Run("Full example", func(t *testing.T) {
			input := func(context.Context, struct {
				PathValue1  string `api:"path:pathkey1" description:"my description"`
				PathValue2  int    `api:"path:pathkey2" description:"my description"`
				QueryValue1 string `api:"query:querykey1" description:"my description"`
				QueryValue2 int    `api:"query:querykey2" description:"my description"`
				Body        struct {
					Value1 string `json:"value1"`
				} `api:"body"`
			}) (*struct{}, error) {
				return nil, nil
			}
			output, err := ParseRestfulFunction(input)
			require.Nil(t, err)
			assert.NotNil(t, output)
			assert.NotNil(t, output.BodyExample)
			assert.Equal(t, 0, output.InContextPosition)
			assert.Equal(t, 1, output.InMetadataPosition)
			assert.NotNil(t, output.InMetadataType)
			assert.Equal(t, 1, output.OutErrorPosition)
			assert.Equal(t, 0, output.OutResponsePosition)
			assert.NotNil(t, output.ResponseExample)

			if assert.Equal(t, 5, len(output.InputFields)) {
				assert.Equal(t, "PathValue1", output.InputFields[0].Name)
				assert.Equal(t, "PathValue2", output.InputFields[1].Name)
				assert.Equal(t, "QueryValue1", output.InputFields[2].Name)
				assert.Equal(t, "QueryValue2", output.InputFields[3].Name)
				assert.Equal(t, "Body", output.InputFields[4].Name)
			}
			if assert.Equal(t, 2, len(output.PathParameters)) {
				assert.Equal(t, "PathValue1", output.PathParameters[0].FieldName)
				assert.Equal(t, "pathkey1", output.PathParameters[0].Name)
				assert.Equal(t, "my description", output.PathParameters[0].Description)

				assert.Equal(t, "PathValue2", output.PathParameters[1].FieldName)
				assert.Equal(t, "pathkey2", output.PathParameters[1].Name)
				assert.Equal(t, "my description", output.PathParameters[1].Description)
			}
			if assert.Equal(t, 2, len(output.QueryParameters)) {
				assert.Equal(t, "QueryValue1", output.QueryParameters[0].FieldName)
				assert.Equal(t, "querykey1", output.QueryParameters[0].Name)
				assert.Equal(t, "my description", output.QueryParameters[0].Description)

				assert.Equal(t, "QueryValue2", output.QueryParameters[1].FieldName)
				assert.Equal(t, "querykey2", output.QueryParameters[1].Name)
				assert.Equal(t, "my description", output.QueryParameters[1].Description)
			}

			f := output.CreateFunctionWithError(nil)
			assert.NotNil(t, f)
		})
	})
}
