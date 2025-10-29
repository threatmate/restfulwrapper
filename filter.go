package restfulwrapper

import (
	"fmt"
	"log/slog"

	"github.com/emicklei/go-restful/v3"
)

// filterSetAttributes sets any attributes on the request.
func filterSetAttributes(attributes map[string]any) func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		defer chain.ProcessFilter(req, resp)

		slog.DebugContext(req.Request.Context(), fmt.Sprintf("Setting %d attributes on request...", len(attributes)))
		for key, value := range attributes {
			req.SetAttribute(key, value)
		}
	}
}
