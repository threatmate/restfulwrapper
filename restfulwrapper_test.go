package restfulwrapper_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/emicklei/go-restful/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/threatmate/restfulwrapper"
)

type API struct {
	_ SubAPI `api:"httppath:/subapi"`
}

type GetEndpoint1Metadata struct {
	restfulwrapper.HTTPMethodGET
	_ string `api:"httppath:/endpoint1"`
	_ string `api:"doc" description:"Endpoint 1 doc."`
	_ string `api:"notes" description:"Endpoint 1 notes"`
}

func (a *API) GetEndpoint1(ctx context.Context, meta GetEndpoint1Metadata) (string, error) {
	return "endpoint1", nil
}

type DeleteEndpoint1Metadata struct {
	restfulwrapper.HTTPMethodDELETE
	_ string `api:"httppath:/endpoint1"`
	_ string `api:"doc" description:"Endpoint 1 doc."`
	_ string `api:"notes" description:"Endpoint 1 notes"`
}

func (a *API) DeleteEndpoint1(ctx context.Context, meta DeleteEndpoint1Metadata) error {
	return nil
}

type SubAPI struct {
}

type PostEndpoint2Metadata struct {
	restfulwrapper.HTTPMethodPOST
	_    string            `api:"httppath:/endpoint2/{id}"`
	_    string            `api:"doc" description:"Endpoint 2 doc."`
	_    string            `api:"notes" description:"Endpoint 2 notes"`
	ID   int               `api:"path:id" description:"ID parameter."`
	Body map[string]string `api:"body" description:"Request body."`
}

func (a *SubAPI) PostEndpoint2(ctx context.Context, meta PostEndpoint2Metadata) (string, error) {
	if meta.Body == nil {
		return "", restfulwrapper.NewAPIBodyError(fmt.Errorf("body is nil"))
	}
	if meta.Body["key1"] == "" {
		return "", restfulwrapper.NewAPIBodyError(fmt.Errorf("key1 is required in body"))
	}
	return fmt.Sprintf("endpoint2:%d:%s", meta.ID, meta.Body["key1"]), nil
}

type CustomError struct{}

var _ error = (*CustomError)(nil)
var _ restfulwrapper.ErrorWriter = (*CustomError)(nil)

func (e *CustomError) Error() string {
	return "custom Error"
}

type GetEndpoint3Metadata struct {
	restfulwrapper.HTTPMethodGET
	_ string `api:"httppath:/endpoint3"`
	_ string `api:"doc" description:"Endpoint 3 doc."`
	_ string `api:"notes" description:"Endpoint 3 notes"`
}

func (a *SubAPI) GetEndpoint3(ctx context.Context, meta GetEndpoint3Metadata) (string, error) {
	return "", fmt.Errorf("wrap3: %w", fmt.Errorf("wrap2: %w", fmt.Errorf("wrap1: %w", fmt.Errorf("some error"))))
}

func (e *CustomError) WriteError(resp *restful.Response) {
	resp.Header().Set("X-Custom-Error", "my custom value")
	resp.Write([]byte(`custom WriteError`))
}

type GetEndpoint4Metadata struct {
	restfulwrapper.HTTPMethodGET
	_ string `api:"httppath:/endpoint4"`
	_ string `api:"doc" description:"Endpoint 4 doc."`
	_ string `api:"notes" description:"Endpoint 4 notes"`
}

func (a *SubAPI) GetEndpoint4(ctx context.Context, meta GetEndpoint4Metadata) (string, error) {
	return "", fmt.Errorf("wrap3: %w", fmt.Errorf("wrap2: %w", fmt.Errorf("wrap1: %w", &CustomError{})))
}

func TestRestfulWrapper(t *testing.T) {
	if value := os.Getenv("DEBUG"); value == "1" || value == "true" {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	}

	ctx := t.Context()

	webService := restfulwrapper.WebService("/api").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	{
		session := webService.Session()
		session.Register(ctx, "/v1", &API{})
	}

	container := restful.NewContainer()
	container.Add(webService.WebService())

	server := httptest.NewServer(container)
	defer server.Close()

	t.Run("GET /api/v1/bogus", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/api/v1/bogus", nil)
		require.Nil(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.Nil(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
	t.Run("GET /api/v1/endpoint1", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/api/v1/endpoint1", nil)
		require.Nil(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.Nil(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		require.Nil(t, err)
		require.Equal(t, `"endpoint1"`, string(bodyBytes))
	})
	t.Run("DELETE /api/v1/endpoint1", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, server.URL+"/api/v1/endpoint1", nil)
		require.Nil(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.Nil(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		require.Nil(t, err)
		require.Empty(t, bodyBytes)
	})
	t.Run("POST /api/v1/endpoint1", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, server.URL+"/api/v1/endpoint1", nil)
		require.Nil(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.Nil(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		require.Nil(t, err)
		require.Equal(t, "405: Method Not Allowed", string(bodyBytes))
	})
	t.Run("POST /api/v1/subapi/endpoint2", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, server.URL+"/api/v1/subapi/endpoint2", nil)
		require.Nil(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.Nil(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
	t.Run("POST /api/v1/subapi/endpoint2/bogus with no content-type", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, server.URL+"/api/v1/subapi/endpoint2/bogus", nil)
		require.Nil(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.Nil(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusUnsupportedMediaType, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		require.Nil(t, err)
		require.Equal(t, "415: Unsupported Media Type", string(bodyBytes))
	})
	t.Run("POST /api/v1/subapi/endpoint2/bogus with bad path", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, server.URL+"/api/v1/subapi/endpoint2/bogus", nil)
		require.Nil(t, err)

		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.Nil(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		require.Nil(t, err)

		var output map[string]string
		err = json.Unmarshal(bodyBytes, &output)
		require.Nil(t, err)
		assert.Equal(t, `*restfulwrapper.APIPathParameterError`, output["type"])
		assert.Equal(t, `strconv.ParseInt: parsing "bogus": invalid syntax`, output["message"])
		assert.Equal(t, `id`, output["parameter"])
	})
	t.Run("POST /api/v1/subapi/endpoint2/1 without body key", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, server.URL+"/api/v1/subapi/endpoint2/1", nil)
		require.Nil(t, err)

		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.Nil(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		require.Nil(t, err)

		var output map[string]string
		err = json.Unmarshal(bodyBytes, &output)
		require.Nil(t, err)
		assert.Equal(t, `*restfulwrapper.APIBodyError`, output["type"])
		assert.Contains(t, output["message"], `EOF`)
		assert.NotContains(t, output, "parameter")
	})
	t.Run("POST /api/v1/subapi/endpoint2/1 with body key", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, server.URL+"/api/v1/subapi/endpoint2/1", strings.NewReader(`{"key1":"value1"}`))
		require.Nil(t, err)

		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.Nil(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		require.Nil(t, err)

		require.Equal(t, `"endpoint2:1:value1"`, string(bodyBytes))
	})
	t.Run("GET /api/v1/subapi/endpoint3", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/api/v1/subapi/endpoint3", nil)
		require.Nil(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.Nil(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		require.Nil(t, err)

		var output map[string]string
		err = json.Unmarshal(bodyBytes, &output)
		require.Nil(t, err)
		assert.Equal(t, `*fmt.wrapError`, output["type"])
		assert.Equal(t, `wrap3: wrap2: wrap1: some error`, output["message"])
		assert.NotContains(t, output, "parameter")
	})
	t.Run("GET /api/v1/subapi/endpoint4", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/api/v1/subapi/endpoint4", nil)
		require.Nil(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.Nil(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		require.Nil(t, err)

		require.Equal(t, `custom WriteError`, string(bodyBytes))
	})
}
