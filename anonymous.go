package restfulwrapper

// HTTPMethodDELETE marks this endpoint as DELETE.
type HTTPMethodDELETE struct {
	_ string `api:"httpmethod:DELETE"`
}

// HTTPMethodGET marks this endpoint as GET.
type HTTPMethodGET struct {
	_ string `api:"httpmethod:GET"`
}

// HTTPMethodOPTIONS marks this endpoint as OPTIONS.
type HTTPMethodOPTIONS struct {
	_ string `api:"httpmethod:OPTIONS"`
}

// HTTPMethodPATCH marks this endpoint as PATCH.
type HTTPMethodPATCH struct {
	_ string `api:"httpmethod:PATCH"`
}

// HTTPMethodPOST marks this endpoint as POST.
type HTTPMethodPOST struct {
	_ string `api:"httpmethod:POST"`
}

// HTTPMethodPUT marks this endpoint as PUT.
type HTTPMethodPUT struct {
	_ string `api:"httpmethod:PUT"`
}
