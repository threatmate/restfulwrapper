package restfulwrapper

import (
	"fmt"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/emicklei/go-restful/v3"
)

// RestfulFunctionInfo contains all of the information about a method that can
// be used as an endpoint.
type RestfulFunctionInfo struct {
	FunctionValue       reflect.Value // This is the function that will be called.
	InContextPosition   int           // This is the position of the context parameter of the method, if any.
	InMetadataPosition  int           // This is the position of the metadata parameter of the method, if any.
	InMetadataType      reflect.Type  // This is the type of the metadata parameter of the method, if any.
	OutErrorPosition    int           // This is the position of the error return value, if any.
	OutResponsePosition int           // This is the position of the response return value, if any.

	HTTPMethod       string                           // This is the HTTP method.
	HTTPPath         string                           // This is the path (including any "{}" router syntax).
	Doc              string                           // Used with "restful".
	Notes            string                           // Used with "restful".
	PathParameters   []RestfulFunctionPathParameter   // Used with "restful".
	QueryParameters  []RestfulFunctionQueryParameter  // Used with "restful".
	HeaderParameters []RestfulFunctionHeaderParameter // Used with "restful".
	BodyExample      any                              // Used with "restful".
	ResponseExample  any                              // Used with "restful".
	Do               []func(*restful.RouteBuilder)    // Used with "restful"; these will be called as "Do" functions.
	Consumes         []string                         // Used with "restful".
	Produces         []string                         // Used with "restful".

	InputFields []InputField // This is the list of fields in the metadata struct and how we populate them.

	LocalMap map[string]string // This is an arbitrary mapping that can be used to store information.
}

// InputField represents a field on the metadata struct.
type InputField struct {
	Name     string             // This is the name of the field.
	Function InputFieldFunction // This is the function that we will call to set its value.
}

// InputFieldFunction sets the value of the field.
type InputFieldFunction func(v reflect.Value, req *restful.Request, metadataValue reflect.Value) error

// RestfulFunctionPathParameter represents a path parameter.
type RestfulFunctionPathParameter struct {
	FieldName   string
	Name        string
	Description string
}

// RestfulFunctionQueryParameter represents a query parameter.
type RestfulFunctionQueryParameter struct {
	FieldName     string
	Name          string
	Description   string
	AllowMultiple bool
}

// RestfulFunctionHeaderParameter represents a header parameter.
type RestfulFunctionHeaderParameter struct {
	FieldName     string
	Name          string
	Description   string
	AllowMultiple bool
}

// UpdateRouteBuilder updates a restful.Routebuilder with the information that we got from
// parsing the function.
func (info *RestfulFunctionInfo) UpdateRouteBuilder(routeBuilder *restful.RouteBuilder) {
	for _, headerParameter := range info.HeaderParameters {
		parameter := restful.HeaderParameter(headerParameter.Name, headerParameter.Description)
		parameter.AllowMultiple(headerParameter.AllowMultiple)
		if headerParameter.AllowMultiple {
			parameter.CollectionFormat(restful.CollectionFormatMulti)
		}
		routeBuilder.Param(parameter)
		routeBuilder.Returns(http.StatusBadRequest, "Bad Request", nil)
	}
	for _, pathParameter := range info.PathParameters {
		parameter := restful.PathParameter(pathParameter.Name, pathParameter.Description)
		parameter.AllowEmptyValue(false)
		routeBuilder.Param(parameter)
		routeBuilder.Returns(http.StatusBadRequest, "Bad Request", nil)
	}
	for _, queryParameter := range info.QueryParameters {
		parameter := restful.QueryParameter(queryParameter.Name, queryParameter.Description)
		parameter.AllowMultiple(queryParameter.AllowMultiple)
		if queryParameter.AllowMultiple {
			parameter.CollectionFormat(restful.CollectionFormatMulti)
		}
		routeBuilder.Param(parameter)
		routeBuilder.Returns(http.StatusBadRequest, "Bad Request", nil)
	}

	if len(info.Consumes) > 0 {
		routeBuilder.Consumes(info.Consumes...)
	}
	if len(info.Produces) > 0 {
		routeBuilder.Produces(info.Produces...)
	}

	if info.BodyExample != nil {
		// We have a body to read.
		routeBuilder.Reads(info.BodyExample)
		// And because we have a body to read, we can fail with a bad request.
		routeBuilder.Returns(http.StatusBadRequest, "Bad Request", nil)
	}
	if info.ResponseExample != nil {
		routeBuilder.Returns(http.StatusOK, "OK", info.ResponseExample)
	}

	routeBuilder.Doc(info.Doc)
	routeBuilder.Notes(info.Notes)

	routeBuilder.Do(info.Do...)
}

// CreateFunctionWithError returns a `RestfulFunctionWithError` using the given attributes.
func (info *RestfulFunctionInfo) CreateFunctionWithError(errorHandler ErrorHandler) RestfulFunctionWithError {
	// Create the function that we'll return.
	functionWithError := func(req *restful.Request, resp *restful.Response) error {
		ctx := req.Request.Context()

		// Create the list of arguments to pass to the method.
		methodArguments := make([]reflect.Value, info.FunctionValue.Type().NumIn())

		// If we have a context that we need to pass in, then set that up.
		if info.InContextPosition >= 0 {
			contextValue := reflect.ValueOf(ctx)
			methodArguments[info.InContextPosition] = contextValue
		}

		// If we have a metadata struct to pass in, then set that up.
		if info.InMetadataPosition >= 0 {
			inputValue := reflect.New(info.FunctionValue.Type().In(info.InMetadataPosition)).Elem()

			argumentType := info.FunctionValue.Type().In(info.InMetadataPosition)
			switch argumentType.Kind() {
			case reflect.Pointer:
				argumentType = argumentType.Elem()
			}

			if argumentType.Kind() != reflect.Struct {
				return fmt.Errorf("unexpected input type: %v", argumentType.Kind())
			}

			for _, inputField := range info.InputFields {
				fieldValue := inputValue.FieldByName(inputField.Name)

				err := inputField.Function(fieldValue, req, inputValue)
				if err != nil {
					return err
				}
			}

			slog.DebugContext(ctx, fmt.Sprintf("Input: %+v", inputValue.Interface()))
			methodArguments[info.InMetadataPosition] = inputValue
		}

		// Call the method.
		methodResults := info.FunctionValue.Call(methodArguments)

		// Sanity check: make sure that the results are what we think they should be.
		if len(methodResults) != info.FunctionValue.Type().NumOut() {
			return fmt.Errorf("unexpected output count: got %d, expected %d", len(methodResults), info.FunctionValue.Type().NumOut())
		}

		// This is the error from the method call.
		var err error
		// If we have an error output, then use that.
		if info.OutErrorPosition >= 0 {
			if methodResults[info.OutErrorPosition].Interface() != nil {
				err = methodResults[info.OutErrorPosition].Interface().(error)
			}
		}
		// If the method failed, then return that error.
		if err != nil {
			if errorHandler != nil {
				newErr := errorHandler(err)
				if newErr != nil {
					err = newErr
				}
			}
			return err
		}

		// If we have a response output, then use that.
		if info.OutResponsePosition >= 0 {
			output := methodResults[info.OutResponsePosition].Interface()
			if output == nil {
				slog.DebugContext(ctx, "No output given; writing OK with nil.")
				resp.WriteHeaderAndEntity(http.StatusOK, nil)
			} else if writer, ok := output.(Writer); ok {
				slog.DebugContext(ctx, "Custom output writer given; calling Write on it.")
				writer.Write(resp)
			} else {
				slog.DebugContext(ctx, "Standard struct given; writing OK with it.")
				resp.WriteHeaderAndEntity(http.StatusOK, output)
			}
		} else {
			slog.DebugContext(ctx, "No output position configured; writing OK with nil.")
			resp.WriteHeaderAndEntity(http.StatusOK, nil)
		}

		return nil
	}

	return functionWithError
}
