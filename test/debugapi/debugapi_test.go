package debugapi_test

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"testing"

	"github.com/emicklei/go-restful/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/threatmate/restapiclient"
	"github.com/threatmate/restfulwrapper"
)

type API struct{}

type genericOutput struct {
	Proto      string              `json:"proto"`
	RemoteAddr string              `json:"remoteAddr"`
	Method     string              `json:"method"`
	RequestURI string              `json:"requestUri"`
	Header     map[string][]string `json:"header"`
	Trailer    map[string][]string `json:"trailer"`
}

type DeleteMetadata struct {
	restfulwrapper.HTTPMethodDELETE
	_       string        `api:"httppath:/"`
	_       string        `api:"doc" description:"Debug a DELETE request."`
	_       string        `api:"notes" description:""`
	Body    []byte        `api:"body:consumes:*/*;empty"`
	Request *http.Request `api:"httprequest"`
}
type DeleteOutput genericOutput

func (a *API) Delete(ctx context.Context, meta DeleteMetadata) (DeleteOutput, error) {
	result := a.debugRequest(ctx, meta.Request)
	return DeleteOutput(result), nil
}

type GetMetadata struct {
	restfulwrapper.HTTPMethodGET
	_       string        `api:"httppath:/"`
	_       string        `api:"doc" description:"Debug a GET request."`
	_       string        `api:"notes" description:""`
	Body    []byte        `api:"body:consumes:*/*;empty"`
	Request *http.Request `api:"httprequest"`
}
type GetOutput genericOutput

func (a *API) Get(ctx context.Context, meta GetMetadata) (GetOutput, error) {
	result := a.debugRequest(ctx, meta.Request)
	return GetOutput(result), nil
}

type PatchMetadata struct {
	restfulwrapper.HTTPMethodPATCH
	_       string        `api:"httppath:/"`
	_       string        `api:"doc" description:"Debug a PATCH request."`
	_       string        `api:"notes" description:""`
	Body    []byte        `api:"body:consumes:*/*;empty"`
	Request *http.Request `api:"httprequest"`
}
type PatchOutput genericOutput

func (a *API) Patch(ctx context.Context, meta PatchMetadata) (PatchOutput, error) {
	result := a.debugRequest(ctx, meta.Request)
	return PatchOutput(result), nil
}

type PostMetadata struct {
	restfulwrapper.HTTPMethodPOST
	_       string        `api:"httppath:/"`
	_       string        `api:"doc" description:"Debug a POST request."`
	_       string        `api:"notes" description:""`
	Body    []byte        `api:"body:consumes:*/*;empty"`
	Request *http.Request `api:"httprequest"`
}
type PostOutput genericOutput

func (a *API) Post(ctx context.Context, meta PostMetadata) (PostOutput, error) {
	result := a.debugRequest(ctx, meta.Request)
	return PostOutput(result), nil
}

type PutMetadata struct {
	restfulwrapper.HTTPMethodPUT
	_       string        `api:"httppath:/"`
	_       string        `api:"doc" description:"Debug a PUT request."`
	_       string        `api:"notes" description:""`
	Body    []byte        `api:"body:consumes:*/*;empty"`
	Request *http.Request `api:"httprequest"`
}
type PutOutput genericOutput

func (a *API) Put(ctx context.Context, meta PutMetadata) (PutOutput, error) {
	result := a.debugRequest(ctx, meta.Request)
	return PutOutput(result), nil
}

// debugRequest returns information about the request.
func (a *API) debugRequest(ctx context.Context, request *http.Request) genericOutput {
	{
		contents, _ := httputil.DumpRequest(request, true)
		slog.DebugContext(ctx, fmt.Sprintf("HTTP request: %s", contents))
	}

	output := genericOutput{
		Proto:      request.Proto,
		RemoteAddr: request.RemoteAddr,
		Method:     request.Method,
		RequestURI: request.RequestURI,
		Header:     request.Header,
		Trailer:    request.Trailer,
	}
	return output
}

func TestDebug(t *testing.T) {
	if value := os.Getenv("DEBUG"); value == "1" || value == "true" {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	}

	ctx := t.Context()

	webService := restfulwrapper.WebService("/api").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	{
		session := webService.Session()
		session.Register(ctx, "/v1/debug/request", &API{})
	}

	container := restful.NewContainer()
	container.Add(webService.WebService())

	server := httptest.NewServer(container)
	defer server.Close()
	slog.DebugContext(ctx, fmt.Sprintf("Debug server listening on: %s", server.URL))

	restClient := restapiclient.New(server.URL)

	t.Run("Debug", func(t *testing.T) {
		methods := []string{
			http.MethodDelete,
			http.MethodGet,
			http.MethodPatch,
			http.MethodPost,
			http.MethodPut,
		}
		for _, method := range methods {
			t.Run(method, func(t *testing.T) {
				input := "test-input-here"
				var output GetOutput
				err := restClient.Do(ctx, method, "/api/v1/debug/request", input, &output)
				require.Nil(t, err)
				assert.Equal(t, method, output.Method)
				assert.Equal(t, "HTTP/1.1", output.Proto)
				assert.Equal(t, []string{fmt.Sprintf("%d", len(`"`+input+`"`))}, output.Header["Content-Length"])
			})
		}
	})
}
