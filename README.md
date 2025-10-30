# restfulwrapper
![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/threatmate/restfulwrapper?label=version&logo=version&sort=semver)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/threatmate/restfulwrapper)](https://pkg.go.dev/github.com/threatmate/restfulwrapper)


This is a wrapper around [go-restful](https://github.com/emicklei/go-restful) to make it easier to use.

# Usage
This is a simple case of a single API container with a single endpoint.
```
type API struct{}

type GetMetadata struct {
	restfulwrapper.HTTPMethodGET
	_ string `api:"httppath:/"`
	_ string `api:"doc" description:"Handle a GET request."`
	_ string `api:"notes" description:""`
}
type GetOutput struct{}

func (a *API) Get(ctx context.Context, meta GetMetadata) (output GetOutput, err error) {
	output = GetOutput{}

	// ...

	return output, nil
}

// ...

webService := restfulwrapper.WebService("/api").
	Consumes(restful.MIME_JSON).
	Produces(restful.MIME_JSON)
{
	session := webService.Session()
	session.Register(ctx, "/v1/path/to/service", &API{})
}

container := restful.NewContainer()
container.Add(webService.WebService())

// Use container as you would any `http.Handler`.
```

An API struct can embed other API structs using `httppath`:
```
type API struct{
	_ OtherAPI `api:"httppath:/other-api"`
}
```
