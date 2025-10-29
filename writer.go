package restfulwrapper

import "github.com/emicklei/go-restful/v3"

// Writer can be used on an output type to control exactly how a response is rendered.
//
// For example, this can be used to set up a redirect with a "Location" header or render
// a PDF with a custom "Content-Type".
type Writer interface {
	Write(*restful.Response)
}
