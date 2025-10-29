package restfulwrapper

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	"github.com/emicklei/go-restful/v3"
)

// WebService creates a new restful.WebService with the given path, but in
// a wrapper that makes it easy to add routes with common properties.
func WebService(path string) *RestfulWrapper {
	return &RestfulWrapper{
		path: path,
		ws:   new(restful.WebService).Path(path),
	}
}

// RestfulFunctionWithError is a restful.RouteFunction that returns an error.
type RestfulFunctionWithError func(req *restful.Request, resp *restful.Response) error

// restfulFunctionWrapper takes our more structured RestfulFunctionWithError function and returns
// a function that restful can directly use.
func restfulFunctionWrapper(f RestfulFunctionWithError) restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		ctx := req.Request.Context()

		err := f(req, resp)
		if err != nil {
			slog.InfoContext(ctx, fmt.Sprintf("Error performing request: [%T] %v", err, err))

			// If the error is a pointer to an ErrorWriter, use it directly.
			{
				var errorWriter ErrorWriter
				if errors.As(err, &errorWriter) {
					slog.InfoContext(ctx, "Error is a pointer to an ErrorWriter; using its custom writer function.")

					errorWriter.WriteError(resp)
					return
				}
			}

			slog.InfoContext(ctx, "Error does not implement ErrorWriter; writing a generic error.")

			output := APIResponseErrorOutput{
				Type:    fmt.Sprintf("%T", err),
				Message: err.Error(),
			}
			resp.WriteHeaderAndEntity(http.StatusInternalServerError, output)
			return
		}
	}
}

// ContextAction is a context action function.
type ContextAction func(ctx context.Context, info *RestfulFunctionInfo) context.Context

// RestfulWrapper is our restful wrapper.
type RestfulWrapper struct {
	ws             *restful.WebService           // This is the WebService; we need this to create parameters.
	path           string                        // This is the path that was initially provided.
	attributes     map[string]any                // This is a list of any attributes to set for every request.
	doFunctions    []func(*restful.RouteBuilder) // This is a list of any "do" functions.
	consumes       []string                      // This is a list of any MIME types that will be consumed.
	produces       []string                      // This is a list of any MIME types that will be produced.
	contextActions []ContextAction               // This is a list of context actions to take for each request.
}

// Session returns a new session of the wrapper.  Any modifications will not affect
// the original instance and will only apply to new routes added to this session.
func (r *RestfulWrapper) Session() *RestfulWrapper {
	newWrapper := &RestfulWrapper{
		ws:             r.ws,
		path:           r.path,
		attributes:     map[string]any{},
		doFunctions:    []func(*restful.RouteBuilder){},
		consumes:       []string{},
		produces:       []string{},
		contextActions: []ContextAction{},
	}

	for key, value := range r.attributes {
		newWrapper.attributes[key] = value
	}
	newWrapper.doFunctions = append(newWrapper.doFunctions, r.doFunctions...)
	newWrapper.consumes = append(newWrapper.consumes, r.consumes...)
	newWrapper.produces = append(newWrapper.produces, r.produces...)
	newWrapper.contextActions = append(newWrapper.contextActions, r.contextActions...)

	return newWrapper
}

// Attributes sets (adds to) the attributes for any request.
//
// These attributes will be accessible via `restful.Request`'s `Attribute` function.
func (r *RestfulWrapper) Attributes(attributes map[string]any) *RestfulWrapper {
	if r.attributes == nil {
		r.attributes = map[string]any{}
	}

	for key, value := range attributes {
		r.attributes[key] = value
	}

	return r
}

// Consumes sets the content types that will be consumed.
func (r *RestfulWrapper) Consumes(contentTypes ...string) *RestfulWrapper {
	r.consumes = append(r.consumes, contentTypes...)
	return r
}

// Do registers restful.RouteBuilder functions that will apply to all subsequent Route calls.
func (r *RestfulWrapper) Do(doFunctions ...func(*restful.RouteBuilder)) *RestfulWrapper {
	r.doFunctions = append(r.doFunctions, doFunctions...)
	return r
}

// Produces sets the content types that will be produced.
func (r *RestfulWrapper) Produces(contentTypes ...string) *RestfulWrapper {
	r.produces = append(r.produces, contentTypes...)
	return r
}

// Route adds a number of routes to the restful.WebService.
func (r *RestfulWrapper) Route(routeBuilders ...*restful.RouteBuilder) *RestfulWrapper {
	for _, routeBuilder := range routeBuilders {
		r.ws.Route(routeBuilder)
	}
	return r
}

// Method sets the method (for compatibility with restful.Method).
func (r *RestfulWrapper) Method(method string) *RestfulRouteWrapper {
	routeWrapper := &RestfulRouteWrapper{
		ws:     r,
		method: method,
	}
	routeWrapper.Do(r.doFunctions...)
	return routeWrapper
}

// DELETE creates a DELETE request (for compatiblity with restful.DELETE).
func (r *RestfulWrapper) DELETE(path string) *RestfulRouteWrapper {
	return r.Method(http.MethodDelete).Path(path)
}

// GET creates a GET request (for compatiblity with restful.GET).
func (r *RestfulWrapper) GET(path string) *RestfulRouteWrapper {
	return r.Method(http.MethodGet).Path(path)
}

// OPTIONS creates a OPTIONS request (for compatiblity with restful.OPTIONS).
func (r *RestfulWrapper) OPTIONS(path string) *RestfulRouteWrapper {
	return r.Method(http.MethodOptions).Path(path)
}

// PATCH creates a PATCH request (for compatiblity with restful.PATCH).
func (r *RestfulWrapper) PATCH(path string) *RestfulRouteWrapper {
	return r.Method(http.MethodPatch).Path(path)
}

// POST creates a POST request (for compatiblity with restful.POST).
func (r *RestfulWrapper) POST(path string) *RestfulRouteWrapper {
	return r.Method(http.MethodPost).Path(path)
}

// PUT creates a PUT request (for compatiblity with restful.PUT).
func (r *RestfulWrapper) PUT(path string) *RestfulRouteWrapper {
	return r.Method(http.MethodPut).Path(path)
}

// WebService returns the underlying restful.WebService for use with the other restful functions.
func (r *RestfulWrapper) WebService() *restful.WebService {
	return r.ws
}

// RestfulRouteWrapper wraps a route and ultimately will result in a `*restful.RouteBuilder` value.
type RestfulRouteWrapper struct {
	ws                *RestfulWrapper               // This is the parent wrapper of this route.
	method            string                        // This is the method.
	path              string                        // This is the path.
	functionWithError RestfulFunctionWithError      // This is the function to call.
	doFunctions       []func(*restful.RouteBuilder) // This is a list of any "do" functions.
	consumes          []string                      // This is a list of any custom mime types that this consumes.
	produces          []string                      // This is a list of any custom mime types that this produces.
}

// Consumes sets the content types that will be consumed.
func (r *RestfulRouteWrapper) Consumes(contentTypes ...string) *RestfulRouteWrapper {
	r.consumes = append(r.consumes, contentTypes...)
	return r
}

// Do registers restful.RouteBuilder functions that will apply to all subsequent Route calls.
func (r *RestfulRouteWrapper) Do(doFunctions ...func(*restful.RouteBuilder)) *RestfulRouteWrapper {
	r.doFunctions = append(r.doFunctions, doFunctions...)
	return r
}

// Path sets the path.
func (r *RestfulRouteWrapper) Path(path string) *RestfulRouteWrapper {
	r.path = path
	return r
}

// Produces sets the content types that will be produced.
func (r *RestfulRouteWrapper) Produces(contentTypes ...string) *RestfulRouteWrapper {
	r.produces = append(r.produces, contentTypes...)
	return r
}

// RouteBuilder returns a RouteBuilder with everything we know so far.
func (r *RestfulRouteWrapper) RouteBuilder() *restful.RouteBuilder {
	routeBuilder := r.ws.ws.
		Method(r.method).
		Path(r.path).
		To(restfulFunctionWrapper(r.functionWithError)).
		Filter(filterSetAttributes(r.ws.attributes)).
		Do(r.doFunctions...)

	if len(r.consumes) > 0 {
		routeBuilder.Consumes(r.consumes...)
	} else if len(r.ws.consumes) > 0 {
		routeBuilder.Consumes(r.ws.consumes...)
	}
	if len(r.produces) > 0 {
		routeBuilder.Produces(r.produces...)
	} else if len(r.ws.produces) > 0 {
		routeBuilder.Produces(r.ws.produces...)
	}

	return routeBuilder
}

// Register is the next-generation `To` replacement that accepts a struct pointer with a
// collection of methods.  Each method that _can_ be used as an endpoint will be used as
// an endpoint.
//
// The path given will be used as the root for any endpoints.  Note that the RestfulWrapper
// itself may already have its own path root; this new path will be appended to that.
func (r *RestfulWrapper) Register(ctx context.Context, path string, f interface{}) {
	var fValue = reflect.ValueOf(f)

	slog.DebugContext(ctx, fmt.Sprintf("Registering: %s at %s", fValue.Type().String(), path))

	for i := range fValue.NumMethod() {
		methodValue := fValue.Method(i)

		info, err := ParseRestfulFunction(methodValue.Interface())
		if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Could not parse function: %v: %v", fValue.Type().Method(i).Name, err))
			panic(fmt.Errorf("could not parse function: %v: %w", fValue.Type().Method(i).Name, err))
		}

		routePath := "/" + strings.Trim(path, "/")
		if cleanPath := strings.Trim(info.HTTPPath, "/"); cleanPath != "" {
			if !strings.HasSuffix(routePath, "/") {
				routePath += "/"
			}
			routePath += cleanPath
		}
		info.HTTPPath = r.path + routePath // Set HTTPPath to the full path within the web service.

		routeWrapper := r.Method(info.HTTPMethod)
		routeWrapper.Path(routePath)
		routeWrapper.functionWithError = info.CreateFunctionWithError()
		{
			fs := []func(*restful.RouteBuilder){
				func(builder *restful.RouteBuilder) {
					builder.Filter(func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
						ctx := req.Request.Context()
						ctx = r.applyContextActions(ctx, info)
						req.Request = req.Request.WithContext(ctx)
						chain.ProcessFilter(req, resp)
					})
				},
			}
			fs = append(fs, routeWrapper.doFunctions...)
			routeWrapper.doFunctions = fs
		}

		routeBuilder := routeWrapper.RouteBuilder()
		info.UpdateRouteBuilder(routeBuilder)

		slog.DebugContext(ctx, fmt.Sprintf("Registering function: %s at %s %s", fValue.Type().Method(i).Name, routeWrapper.method, routeWrapper.path))
		r.ws.Route(routeBuilder)
	}

	for fValue.Kind() == reflect.Pointer {
		fValue = fValue.Elem()
	}
	for i := range fValue.NumField() {
		fieldValue := fValue.Field(i)

		apiTagValue := fValue.Type().Field(i).Tag.Get("api")
		if len(apiTagValue) > 0 {
			tagParts := strings.Split(apiTagValue, ";")
			for _, tagPart := range tagParts {
				tagPartParts := strings.SplitN(tagPart, ":", 2)
				tagPartKey := tagPartParts[0]
				var tagPartValue string
				if len(tagPartParts) > 1 {
					tagPartValue = tagPartParts[1]
				}

				if tagPartKey == "httppath" {
					var fieldInterface any
					if fieldValue.CanSet() {
						if fieldValue.CanAddr() {
							fieldInterface = fieldValue.Addr().Interface()
						} else {
							fieldInterface = fieldValue.Interface()
						}
					} else {
						newValue := reflect.New(fieldValue.Type())
						fieldInterface = newValue.Interface()
					}
					r.Register(ctx, strings.TrimRight(path, "/")+"/"+strings.TrimLeft(tagPartValue, "/"), fieldInterface)
				}
			}
		}
	}
}

func (w *RestfulWrapper) ContextAction(f ...ContextAction) *RestfulWrapper {
	w.contextActions = append(w.contextActions, f...)
	return w
}

func (w *RestfulWrapper) applyContextActions(ctx context.Context, info *RestfulFunctionInfo) context.Context {
	for _, contextAction := range w.contextActions {
		ctx = contextAction(ctx, info)
	}
	return ctx
}
