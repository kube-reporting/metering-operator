/*
Package restful , a lean package for creating REST-style WebServices without magic.

WebServices and Routes

A WebService has a collection of Route objects that dispatch incoming Http Requests to a function calls.
Typically, a WebService has a root path (e.g. /users) and de***REMOVED***nes common MIME types for its routes.
WebServices must be added to a container (see below) in order to handler Http requests from a server.

A Route is de***REMOVED***ned by a HTTP method, an URL path and (optionally) the MIME types it consumes (Content-Type) and produces (Accept).
This package has the logic to ***REMOVED***nd the best matching Route and if found, call its Function.

	ws := new(restful.WebService)
	ws.
		Path("/users").
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	ws.Route(ws.GET("/{user-id}").To(u.***REMOVED***ndUser))  // u is a UserResource

	...

	// GET http://localhost:8080/users/1
	func (u UserResource) ***REMOVED***ndUser(request *restful.Request, response *restful.Response) {
		id := request.PathParameter("user-id")
		...
	}

The (*Request, *Response) arguments provide functions for reading information from the request and writing information back to the response.

See the example https://github.com/emicklei/go-restful/blob/master/examples/restful-user-resource.go with a full implementation.

Regular expression matching Routes

A Route parameter can be speci***REMOVED***ed using the format "uri/{var[:regexp]}" or the special version "uri/{var:*}" for matching the tail of the path.
For example, /persons/{name:[A-Z][A-Z]} can be used to restrict values for the parameter "name" to only contain capital alphabetic characters.
Regular expressions must use the standard Go syntax as described in the regexp package. (https://code.google.com/p/re2/wiki/Syntax)
This feature requires the use of a CurlyRouter.

Containers

A Container holds a collection of WebServices, Filters and a http.ServeMux for multiplexing http requests.
Using the statements "restful.Add(...) and restful.Filter(...)" will register WebServices and Filters to the Default Container.
The Default container of go-restful uses the http.DefaultServeMux.
You can create your own Container and create a new http.Server for that particular container.

	container := restful.NewContainer()
	server := &http.Server{Addr: ":8081", Handler: container}

Filters

A ***REMOVED***lter dynamically intercepts requests and responses to transform or use the information contained in the requests or responses.
You can use ***REMOVED***lters to perform generic logging, measurement, authentication, redirect, set response headers etc.
In the restful package there are three hooks into the request,response flow where ***REMOVED***lters can be added.
Each ***REMOVED***lter must de***REMOVED***ne a FilterFunction:

	func (req *restful.Request, resp *restful.Response, chain *restful.FilterChain)

Use the following statement to pass the request,response pair to the next ***REMOVED***lter or RouteFunction

	chain.ProcessFilter(req, resp)

Container Filters

These are processed before any registered WebService.

	// install a (global) ***REMOVED***lter for the default container (processed before any webservice)
	restful.Filter(globalLogging)

WebService Filters

These are processed before any Route of a WebService.

	// install a webservice ***REMOVED***lter (processed before any route)
	ws.Filter(webserviceLogging).Filter(measureTime)


Route Filters

These are processed before calling the function associated with the Route.

	// install 2 chained route ***REMOVED***lters (processed before calling ***REMOVED***ndUser)
	ws.Route(ws.GET("/{user-id}").Filter(routeLogging).Filter(NewCountFilter().routeCounter).To(***REMOVED***ndUser))

See the example https://github.com/emicklei/go-restful/blob/master/examples/restful-***REMOVED***lters.go with full implementations.

Response Encoding

Two encodings are supported: gzip and deflate. To enable this for all responses:

	restful.DefaultContainer.EnableContentEncoding(true)

If a Http request includes the Accept-Encoding header then the response content will be compressed using the speci***REMOVED***ed encoding.
Alternatively, you can create a Filter that performs the encoding and install it per WebService or Route.

See the example https://github.com/emicklei/go-restful/blob/master/examples/restful-encoding-***REMOVED***lter.go

OPTIONS support

By installing a pre-de***REMOVED***ned container ***REMOVED***lter, your Webservice(s) can respond to the OPTIONS Http request.

	Filter(OPTIONSFilter())

CORS

By installing the ***REMOVED***lter of a CrossOriginResourceSharing (CORS), your WebService(s) can handle CORS requests.

	cors := CrossOriginResourceSharing{ExposeHeaders: []string{"X-My-Header"}, CookiesAllowed: false, Container: DefaultContainer}
	Filter(cors.Filter)

Error Handling

Unexpected things happen. If a request cannot be processed because of a failure, your service needs to tell via the response what happened and why.
For this reason HTTP status codes exist and it is important to use the correct code in every exceptional situation.

	400: Bad Request

If path or query parameters are not valid (content or type) then use http.StatusBadRequest.

	404: Not Found

Despite a valid URI, the resource requested may not be available

	500: Internal Server Error

If the application logic could not process the request (or write the response) then use http.StatusInternalServerError.

	405: Method Not Allowed

The request has a valid URL but the method (GET,PUT,POST,...) is not allowed.

	406: Not Acceptable

The request does not have or has an unknown Accept Header set for this operation.

	415: Unsupported Media Type

The request does not have or has an unknown Content-Type Header set for this operation.

ServiceError

In addition to setting the correct (error) Http status code, you can choose to write a ServiceError message on the response.

Performance options

This package has several options that affect the performance of your service. It is important to understand them and how you can change it.

	restful.DefaultContainer.DoNotRecover(false)

DoNotRecover controls whether panics will be caught to return HTTP 500.
If set to false, the container will recover from panics.
Default value is true

	restful.SetCompressorProvider(NewBoundedCachedCompressors(20, 20))

If content encoding is enabled then the default strategy for getting new gzip/zlib writers and readers is to use a sync.Pool.
Because writers are expensive structures, performance is even more improved when using a preloaded cache. You can also inject your own implementation.

Trouble shooting

This package has the means to produce detail logging of the complete Http request matching process and ***REMOVED***lter invocation.
Enabling this feature requires you to set an implementation of restful.StdLogger (e.g. log.Logger) instance such as:

	restful.TraceLogger(log.New(os.Stdout, "[restful] ", log.LstdFlags|log.Lshort***REMOVED***le))

Logging

The restful.SetLogger() method allows you to override the logger used by the package. By default restful
uses the standard library `log` package and logs to stdout. Different logging packages are supported as
long as they conform to `StdLogger` interface de***REMOVED***ned in the `log` sub-package, writing an adapter for your
preferred package is simple.

Resources

[project]: https://github.com/emicklei/go-restful

[examples]: https://github.com/emicklei/go-restful/blob/master/examples

[design]:  http://ernestmicklei.com/2012/11/11/go-restful-api-design/

[showcases]: https://github.com/emicklei/mora, https://github.com/emicklei/landskape

(c) 2012-2015, http://ernestmicklei.com. MIT License
*/
package restful
