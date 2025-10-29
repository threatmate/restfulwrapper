package restfulwrapper

import (
	"fmt"
	"io"
	"log/slog"
	"reflect"
	"slices"
	"strings"

	"github.com/emicklei/go-restful/v3"
)

func init() {
	// body is used to set the body from a PATCH, POST, or PUT method.
	//
	// Additional fields:
	// * consumes:${content-type}; this sets the content type that is expected.
	// * empty; if true, empty bodies will be allowed.
	//
	// JSON is supported trivially using the "json" package, so you may use a full object here.
	// YAML is supported as "application/x-yaml" using either a "string" or "[]byte" type.
	// HTML forms are supported as "multipart/form-data" using either "multipart.Form" or "*multipart.Form".
	Register("body", func(apiTagValue string, field reflect.StructField, info *RestfulFunctionInfo) (InputFieldFunction, error) {
		var consumes []string
		allowEmpty := false

		if len(apiTagValue) > 0 {
			tagParts := strings.Split(apiTagValue, ";")
			for _, tagPart := range tagParts {
				tagPartParts := strings.SplitN(tagPart, ":", 2)
				tagPartKey := tagPartParts[0]
				var tagPartValue string
				if len(tagPartParts) > 1 {
					tagPartValue = tagPartParts[1]
				}

				switch tagPartKey {
				case "consumes":
					valueParts := strings.Split(tagPartValue, ",")
					for _, valuePart := range valueParts {
						consumes = append(consumes, strings.TrimSpace(valuePart))
					}
				case "empty":
					if tagPartValue != "" {
						return nil, fmt.Errorf("invalid body tag value for empty: %s", tagPartValue)
					}
					allowEmpty = true
				default:
					return nil, fmt.Errorf("invalid body tag: %s", tagPartKey)
				}
			}
		}

		if !allowEmpty {
			exampleValue := reflect.New(field.Type)
			if exampleValue.Kind() == reflect.Pointer {
				exampleValue = exampleValue.Elem()
			}
			info.BodyExample = exampleValue.Interface()
		}
		info.Consumes = consumes

		// If the field type is one of the special ones that we know how to support, then
		// handle them appropriately (set the content type).
		switch field.Type.String() {
		case "url.Values", "*url.Values":
			if len(info.Consumes) == 0 {
				info.Consumes = append(info.Consumes, "application/x-www-form-urlencoded")
			}
		case "multipart.Form", "*multipart.Form":
			if len(info.Consumes) == 0 {
				info.Consumes = append(info.Consumes, "multipart/form-data")
			}
		default:
			// Don't do anything special; we'll use "ReadEntity" later.
		}

		// If the content type is "application/x-www-form-urlencoded", then fail if the field type is incorrect.
		if slices.Contains(info.Consumes, "application/x-www-form-urlencoded") {
			switch field.Type.String() {
			case "url.Values":
			case "*url.Values":
			default:
				return nil, fmt.Errorf("invalid type for content-type application/x-www-form-urlencoded: %s", field.Type.String())
			}
		}
		// If the content type is "multipart/form-data", then fail if the field type is incorrect.
		if slices.Contains(info.Consumes, "multipart/form-data") {
			switch field.Type.String() {
			case "multipart.Form":
			case "*multipart.Form":
			default:
				return nil, fmt.Errorf("invalid type for content-type multipart/form-data: %s", field.Type.String())
			}
		}

		return func(v reflect.Value, req *restful.Request, metadataValue reflect.Value) error {
			if !allowEmpty {
				v.Set(reflect.New(field.Type).Elem())

				contentType := ""
				if len(info.Consumes) > 0 {
					contentType = info.Consumes[0]
				}
				slog.DebugContext(req.Request.Context(), fmt.Sprintf("Content-Type: %s", contentType))

				switch contentType {
				case "application/x-www-form-urlencoded":
					err := req.Request.ParseForm()
					if err != nil {
						return NewAPIBodyError(err)
					}

					switch field.Type.String() {
					case "url.Values":
						v.Set(reflect.ValueOf(req.Request.PostForm))
					case "*url.Values":
						v.Set(reflect.ValueOf(&req.Request.PostForm))
					}
				case "multipart/form-data":
					multipartReader, err := req.Request.MultipartReader()
					if err != nil {
						return NewAPIBodyError(err)
					}
					multipartForm, err := multipartReader.ReadForm(10 * 1000 * 1000 /*10MB in RAM*/)
					if err != nil {
						return NewAPIBodyError(err)
					}

					switch field.Type.String() {
					case "multipart.Form":
						v.Set(reflect.ValueOf(*multipartForm))
					case "*multipart.Form":
						v.Set(reflect.ValueOf(multipartForm))
					}
				default:
					// If they asked for a string, then read the body as a string.
					if v.Kind() == reflect.String {
						contents, err := io.ReadAll(req.Request.Body)
						if err != nil {
							return NewAPIBodyError(err)
						}
						v.Set(reflect.ValueOf(string(contents)))
						return nil
					}

					// If they asked for a byte slice, then read the body as a byte slice.
					if v.Type().String() == "[]byte" {
						contents, err := io.ReadAll(req.Request.Body)
						if err != nil {
							return NewAPIBodyError(err)
						}
						v.Set(reflect.ValueOf(contents))
						return nil
					}

					// Otherwise, attempt to use restful's default method.
					err := req.ReadEntity(v.Addr().Interface())
					if err != nil {
						return NewAPIBodyError(err)
					}
				}
			}
			return nil
		}, nil
	})
	Register("doc", func(apiTagValue string, field reflect.StructField, info *RestfulFunctionInfo) (InputFieldFunction, error) {
		if apiTagValue != "" {
			return nil, fmt.Errorf("unexpected tag value: %s", apiTagValue)
		}

		switch field.Type.Kind() {
		case reflect.String:
		default:
			return nil, fmt.Errorf("bad kind: %s", field.Type.Kind().String())
		}

		info.Doc = field.Tag.Get("description")

		return func(v reflect.Value, req *restful.Request, metadataValue reflect.Value) error {
			if v.CanSet() {
				v.SetString(info.Doc)
			}
			return nil
		}, nil
	})
	Register("header", func(apiTagValue string, field reflect.StructField, info *RestfulFunctionInfo) (InputFieldFunction, error) {
		if apiTagValue == "" {
			return nil, fmt.Errorf("missing tag value")
		}
		if slices.ContainsFunc(info.HeaderParameters, func(item RestfulFunctionHeaderParameter) bool { return item.Name == apiTagValue }) {
			return nil, fmt.Errorf("duplicate header tag")
		}
		info.HeaderParameters = append(info.HeaderParameters, RestfulFunctionHeaderParameter{
			FieldName:   field.Name,
			Name:        apiTagValue,
			Description: field.Tag.Get("description"),
		})
		return func(v reflect.Value, req *restful.Request, metadataValue reflect.Value) error {
			ctx := req.Request.Context()

			stringValue := req.HeaderParameter(apiTagValue)

			err := parseStringToSingleValue(stringValue, v.Addr().Interface())
			if err != nil {
				return NewAPIHeaderParameterError(apiTagValue, err)
			}
			slog.DebugContext(ctx, fmt.Sprintf("header: %s: Parsed %q to %+v.", apiTagValue, stringValue, v.Interface()))
			return nil
		}, nil
	})
	Register("httpmethod", func(apiTagValue string, field reflect.StructField, info *RestfulFunctionInfo) (InputFieldFunction, error) {
		if apiTagValue == "" {
			return nil, fmt.Errorf("missing tag value: %s", apiTagValue)
		}

		switch field.Type.Kind() {
		case reflect.String:
		default:
			return nil, fmt.Errorf("bad kind: %s", field.Type.Kind().String())
		}

		info.HTTPMethod = apiTagValue

		return func(v reflect.Value, req *restful.Request, metadataValue reflect.Value) error {
			if v.CanSet() {
				v.SetString(req.Request.Method)
			}
			return nil
		}, nil
	})
	Register("httppath", func(apiTagValue string, field reflect.StructField, info *RestfulFunctionInfo) (InputFieldFunction, error) {
		if apiTagValue == "" {
			return nil, fmt.Errorf("missing tag value: %s", apiTagValue)
		}

		switch field.Type.Kind() {
		case reflect.String:
		default:
			return nil, fmt.Errorf("bad kind: %s", field.Type.Kind().String())
		}

		info.HTTPPath = apiTagValue

		return func(v reflect.Value, req *restful.Request, metadataValue reflect.Value) error {
			if v.CanSet() {
				v.SetString(req.Request.URL.Path)
			}
			return nil
		}, nil
	})
	Register("httprequest", func(apiTagValue string, field reflect.StructField, info *RestfulFunctionInfo) (InputFieldFunction, error) {
		if apiTagValue != "" {
			return nil, fmt.Errorf("unexpected tag value: %s", apiTagValue)
		}

		switch field.Type.String() {
		case "*http.Request":
		default:
			return nil, fmt.Errorf("bad type: %s", field.Type.String())
		}

		return func(v reflect.Value, req *restful.Request, metadataValue reflect.Value) error {
			if v.CanSet() {
				v.Set(reflect.ValueOf(req.Request))
			}
			return nil
		}, nil
	})
	Register("notes", func(apiTagValue string, field reflect.StructField, info *RestfulFunctionInfo) (InputFieldFunction, error) {
		if apiTagValue != "" {
			return nil, fmt.Errorf("unexpected tag value: %s", apiTagValue)
		}

		switch field.Type.Kind() {
		case reflect.String:
		default:
			return nil, fmt.Errorf("bad kind: %s", field.Type.Kind().String())
		}

		info.Notes = field.Tag.Get("description")

		return func(v reflect.Value, req *restful.Request, metadataValue reflect.Value) error {
			if v.CanSet() {
				v.SetString(info.Notes)
			}
			return nil
		}, nil
	})
	Register("path", func(apiTagValue string, field reflect.StructField, info *RestfulFunctionInfo) (InputFieldFunction, error) {
		if apiTagValue == "" {
			return nil, fmt.Errorf("missing tag value")
		}
		if slices.ContainsFunc(info.PathParameters, func(item RestfulFunctionPathParameter) bool { return item.Name == apiTagValue }) {
			return nil, fmt.Errorf("duplicate path tag")
		}
		info.PathParameters = append(info.PathParameters, RestfulFunctionPathParameter{
			FieldName:   field.Name,
			Name:        apiTagValue,
			Description: field.Tag.Get("description"),
		})
		return func(v reflect.Value, req *restful.Request, metadataValue reflect.Value) error {
			ctx := req.Request.Context()

			stringValue := req.PathParameter(apiTagValue)

			err := parseStringToSingleValue(stringValue, v.Addr().Interface())
			if err != nil {
				return NewAPIPathParameterError(apiTagValue, err)
			}
			slog.DebugContext(ctx, fmt.Sprintf("path: %s: Parsed %q to %+v.", apiTagValue, stringValue, v.Interface()))
			return nil
		}, nil
	})
	Register("produces", func(apiTagValue string, field reflect.StructField, info *RestfulFunctionInfo) (InputFieldFunction, error) {
		if apiTagValue == "" {
			return nil, fmt.Errorf("missing tag value")
		}

		switch field.Type.String() {
		case "string":
		default:
			return nil, fmt.Errorf("bad type: %s", field.Type.String())
		}

		info.Produces = append(info.Produces, apiTagValue)

		return func(v reflect.Value, req *restful.Request, metadataValue reflect.Value) error {
			if v.CanSet() {
				v.Set(reflect.ValueOf(apiTagValue))
			}
			return nil
		}, nil
	})

	Register("query", func(apiTagValue string, field reflect.StructField, info *RestfulFunctionInfo) (InputFieldFunction, error) {
		if apiTagValue == "" {
			return nil, fmt.Errorf("missing tag value")
		}
		names := strings.Split(apiTagValue, ",")
		for i := range names {
			names[i] = strings.TrimSpace(names[i])
		}
		primaryName := names[0]
		for _, name := range names {
			if slices.ContainsFunc(info.QueryParameters, func(item RestfulFunctionQueryParameter) bool { return item.Name == name }) {
				return nil, fmt.Errorf("duplicate query tag: %s", name)
			}
		}
		info.QueryParameters = append(info.QueryParameters, RestfulFunctionQueryParameter{
			FieldName:     field.Name,
			Name:          primaryName,
			Description:   field.Tag.Get("description"),
			AllowMultiple: field.Type.Kind() == reflect.Slice,
		})
		for _, name := range names[1:] {
			info.QueryParameters = append(info.QueryParameters, RestfulFunctionQueryParameter{
				FieldName:     field.Name,
				Name:          name,
				Description:   fmt.Sprintf(`Deprecated; use "%s" instead.`, primaryName),
				AllowMultiple: field.Type.Kind() == reflect.Slice,
			})
		}
		return func(v reflect.Value, req *restful.Request, metadataValue reflect.Value) error {
			ctx := req.Request.Context()

			var name string           // This is the name of the parameter that was used.
			var stringValues []string // This is the list of values for the parameter.
			// Figure out which parameter was used.
			// (Stop once we find a matching parameter.)
			for _, n := range names {
				v := req.QueryParameters(n)
				if len(v) > 0 {
					name = n
					stringValues = v
					break // Stop here; we matched.
				}
			}
			if len(stringValues) == 0 {
				if defaultValue, hasDefault := field.Tag.Lookup("default"); hasDefault {
					stringValues = []string{defaultValue}
				}
			}
			if v.Kind() == reflect.Slice {
				v.Set(reflect.MakeSlice(v.Type(), len(stringValues), len(stringValues)))

				for stringValueIndex, stringValue := range stringValues {
					sliceItem := v.Index(stringValueIndex)

					var queryValue any
					if sliceItem.Kind() == reflect.Pointer {
						sliceItem.Set(reflect.New(sliceItem.Type().Elem()))
						queryValue = sliceItem.Interface()
					} else {
						queryValue = sliceItem.Addr().Interface()
					}

					err := parseStringToSingleValue(stringValue, queryValue)
					if err != nil {
						return NewAPIQueryParameterError(name, err)
					}
					slog.DebugContext(ctx, fmt.Sprintf("query: %s: Parsed %q to %+v.", name, stringValue, reflect.ValueOf(queryValue).Elem()))
				}
			} else {
				if len(stringValues) > 0 {
					if len(stringValues) > 1 {
						slog.WarnContext(ctx, fmt.Sprintf("Multiple values given for query parameter %s: %v", name, stringValues))
					}

					stringValue := stringValues[0]

					var queryValue any
					if v.Kind() == reflect.Pointer {
						v.Set(reflect.New(v.Type().Elem()))
						queryValue = v.Interface()
					} else {
						queryValue = v.Addr().Interface()
					}

					err := parseStringToSingleValue(stringValue, queryValue)
					if err != nil {
						return NewAPIQueryParameterError(name, err)
					}
					slog.DebugContext(ctx, fmt.Sprintf("query: %s: Parsed %q to %+v.", name, stringValue, reflect.ValueOf(queryValue).Elem()))
				}
			}
			return nil
		}, nil
	})
	Register("restfulrequest", func(apiTagValue string, field reflect.StructField, info *RestfulFunctionInfo) (InputFieldFunction, error) {
		if apiTagValue != "" {
			return nil, fmt.Errorf("unexpected tag value: %s", apiTagValue)
		}

		switch field.Type.String() {
		case "*restful.Request":
		default:
			return nil, fmt.Errorf("bad type: %s", field.Type.String())
		}

		return func(v reflect.Value, req *restful.Request, metadataValue reflect.Value) error {
			if v.CanSet() {
				v.Set(reflect.ValueOf(req))
			}
			return nil
		}, nil
	})
}
