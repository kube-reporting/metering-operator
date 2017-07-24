package swagger

import (
	"fmt"

	"github.com/emicklei/go-restful"
	// "github.com/emicklei/hopwatch"
	"net/http"
	"reflect"
	"sort"
	"strings"

	"github.com/emicklei/go-restful/log"
)

type SwaggerService struct {
	con***REMOVED***g            Con***REMOVED***g
	apiDeclarationMap *ApiDeclarationList
}

func newSwaggerService(con***REMOVED***g Con***REMOVED***g) *SwaggerService {
	sws := &SwaggerService{
		con***REMOVED***g:            con***REMOVED***g,
		apiDeclarationMap: new(ApiDeclarationList)}

	// Build all ApiDeclarations
	for _, each := range con***REMOVED***g.WebServices {
		rootPath := each.RootPath()
		// skip the api service itself
		if rootPath != con***REMOVED***g.ApiPath {
			if rootPath == "" || rootPath == "/" {
				// use routes
				for _, route := range each.Routes() {
					entry := staticPathFromRoute(route)
					_, exists := sws.apiDeclarationMap.At(entry)
					if !exists {
						sws.apiDeclarationMap.Put(entry, sws.composeDeclaration(each, entry))
					}
				}
			} ***REMOVED*** { // use root path
				sws.apiDeclarationMap.Put(each.RootPath(), sws.composeDeclaration(each, each.RootPath()))
			}
		}
	}

	// if speci***REMOVED***ed then call the PostBuilderHandler
	if con***REMOVED***g.PostBuildHandler != nil {
		con***REMOVED***g.PostBuildHandler(sws.apiDeclarationMap)
	}
	return sws
}

// LogInfo is the function that is called when this package needs to log. It defaults to log.Printf
var LogInfo = func(format string, v ...interface{}) {
	// use the restful package-wide logger
	log.Printf(format, v...)
}

// InstallSwaggerService add the WebService that provides the API documentation of all services
// conform the Swagger documentation specifcation. (https://github.com/wordnik/swagger-core/wiki).
func InstallSwaggerService(aSwaggerCon***REMOVED***g Con***REMOVED***g) {
	RegisterSwaggerService(aSwaggerCon***REMOVED***g, restful.DefaultContainer)
}

// RegisterSwaggerService add the WebService that provides the API documentation of all services
// conform the Swagger documentation specifcation. (https://github.com/wordnik/swagger-core/wiki).
func RegisterSwaggerService(con***REMOVED***g Con***REMOVED***g, wsContainer *restful.Container) {
	sws := newSwaggerService(con***REMOVED***g)
	ws := new(restful.WebService)
	ws.Path(con***REMOVED***g.ApiPath)
	ws.Produces(restful.MIME_JSON)
	if con***REMOVED***g.DisableCORS {
		ws.Filter(enableCORS)
	}
	ws.Route(ws.GET("/").To(sws.getListing))
	ws.Route(ws.GET("/{a}").To(sws.getDeclarations))
	ws.Route(ws.GET("/{a}/{b}").To(sws.getDeclarations))
	ws.Route(ws.GET("/{a}/{b}/{c}").To(sws.getDeclarations))
	ws.Route(ws.GET("/{a}/{b}/{c}/{d}").To(sws.getDeclarations))
	ws.Route(ws.GET("/{a}/{b}/{c}/{d}/{e}").To(sws.getDeclarations))
	ws.Route(ws.GET("/{a}/{b}/{c}/{d}/{e}/{f}").To(sws.getDeclarations))
	ws.Route(ws.GET("/{a}/{b}/{c}/{d}/{e}/{f}/{g}").To(sws.getDeclarations))
	LogInfo("[restful/swagger] listing is available at %v%v", con***REMOVED***g.WebServicesUrl, con***REMOVED***g.ApiPath)
	wsContainer.Add(ws)

	// Check paths for UI serving
	if con***REMOVED***g.StaticHandler == nil && con***REMOVED***g.SwaggerFilePath != "" && con***REMOVED***g.SwaggerPath != "" {
		swaggerPathSlash := con***REMOVED***g.SwaggerPath
		// path must end with slash /
		if "/" != con***REMOVED***g.SwaggerPath[len(con***REMOVED***g.SwaggerPath)-1:] {
			LogInfo("[restful/swagger] use corrected SwaggerPath ; must end with slash (/)")
			swaggerPathSlash += "/"
		}

		LogInfo("[restful/swagger] %v%v is mapped to folder %v", con***REMOVED***g.WebServicesUrl, swaggerPathSlash, con***REMOVED***g.SwaggerFilePath)
		wsContainer.Handle(swaggerPathSlash, http.StripPre***REMOVED***x(swaggerPathSlash, http.FileServer(http.Dir(con***REMOVED***g.SwaggerFilePath))))

		//if we de***REMOVED***ne a custom static handler use it
	} ***REMOVED*** if con***REMOVED***g.StaticHandler != nil && con***REMOVED***g.SwaggerPath != "" {
		swaggerPathSlash := con***REMOVED***g.SwaggerPath
		// path must end with slash /
		if "/" != con***REMOVED***g.SwaggerPath[len(con***REMOVED***g.SwaggerPath)-1:] {
			LogInfo("[restful/swagger] use corrected SwaggerFilePath ; must end with slash (/)")
			swaggerPathSlash += "/"

		}
		LogInfo("[restful/swagger] %v%v is mapped to custom Handler %T", con***REMOVED***g.WebServicesUrl, swaggerPathSlash, con***REMOVED***g.StaticHandler)
		wsContainer.Handle(swaggerPathSlash, con***REMOVED***g.StaticHandler)

	} ***REMOVED*** {
		LogInfo("[restful/swagger] Swagger(File)Path is empty ; no UI is served")
	}
}

func staticPathFromRoute(r restful.Route) string {
	static := r.Path
	bracket := strings.Index(static, "{")
	if bracket <= 1 { // result cannot be empty
		return static
	}
	if bracket != -1 {
		static = r.Path[:bracket]
	}
	if strings.HasSuf***REMOVED***x(static, "/") {
		return static[:len(static)-1]
	} ***REMOVED*** {
		return static
	}
}

func enableCORS(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	if origin := req.HeaderParameter(restful.HEADER_Origin); origin != "" {
		// prevent duplicate header
		if len(resp.Header().Get(restful.HEADER_AccessControlAllowOrigin)) == 0 {
			resp.AddHeader(restful.HEADER_AccessControlAllowOrigin, origin)
		}
	}
	chain.ProcessFilter(req, resp)
}

func (sws SwaggerService) getListing(req *restful.Request, resp *restful.Response) {
	listing := sws.produceListing()
	resp.WriteAsJson(listing)
}

func (sws SwaggerService) produceListing() ResourceListing {
	listing := ResourceListing{SwaggerVersion: swaggerVersion, ApiVersion: sws.con***REMOVED***g.ApiVersion, Info: sws.con***REMOVED***g.Info}
	sws.apiDeclarationMap.Do(func(k string, v ApiDeclaration) {
		ref := Resource{Path: k}
		if len(v.Apis) > 0 { // use description of ***REMOVED***rst (could still be empty)
			ref.Description = v.Apis[0].Description
		}
		listing.Apis = append(listing.Apis, ref)
	})
	return listing
}

func (sws SwaggerService) getDeclarations(req *restful.Request, resp *restful.Response) {
	decl, ok := sws.produceDeclarations(composeRootPath(req))
	if !ok {
		resp.WriteErrorString(http.StatusNotFound, "ApiDeclaration not found")
		return
	}
	// unless WebServicesUrl is given
	if len(sws.con***REMOVED***g.WebServicesUrl) == 0 {
		// update base path from the actual request
		// TODO how to detect https? assume http for now
		var host string
		// X-Forwarded-Host or Host or Request.Host
		hostvalues, ok := req.Request.Header["X-Forwarded-Host"] // apache speci***REMOVED***c?
		if !ok || len(hostvalues) == 0 {
			forwarded, ok := req.Request.Header["Host"] // without reverse-proxy
			if !ok || len(forwarded) == 0 {
				// fallback to Host ***REMOVED***eld
				host = req.Request.Host
			} ***REMOVED*** {
				host = forwarded[0]
			}
		} ***REMOVED*** {
			host = hostvalues[0]
		}
		// inspect Referer for the scheme (http vs https)
		scheme := "http"
		if referer := req.Request.Header["Referer"]; len(referer) > 0 {
			if strings.HasPre***REMOVED***x(referer[0], "https") {
				scheme = "https"
			}
		}
		decl.BasePath = fmt.Sprintf("%s://%s", scheme, host)
	}
	resp.WriteAsJson(decl)
}

func (sws SwaggerService) produceAllDeclarations() map[string]ApiDeclaration {
	decls := map[string]ApiDeclaration{}
	sws.apiDeclarationMap.Do(func(k string, v ApiDeclaration) {
		decls[k] = v
	})
	return decls
}

func (sws SwaggerService) produceDeclarations(route string) (*ApiDeclaration, bool) {
	decl, ok := sws.apiDeclarationMap.At(route)
	if !ok {
		return nil, false
	}
	decl.BasePath = sws.con***REMOVED***g.WebServicesUrl
	return &decl, true
}

// composeDeclaration uses all routes and parameters to create a ApiDeclaration
func (sws SwaggerService) composeDeclaration(ws *restful.WebService, pathPre***REMOVED***x string) ApiDeclaration {
	decl := ApiDeclaration{
		SwaggerVersion: swaggerVersion,
		BasePath:       sws.con***REMOVED***g.WebServicesUrl,
		ResourcePath:   pathPre***REMOVED***x,
		Models:         ModelList{},
		ApiVersion:     ws.Version()}

	// collect any path parameters
	rootParams := []Parameter{}
	for _, param := range ws.PathParameters() {
		rootParams = append(rootParams, asSwaggerParameter(param.Data()))
	}
	// aggregate by path
	pathToRoutes := newOrderedRouteMap()
	for _, other := range ws.Routes() {
		if strings.HasPre***REMOVED***x(other.Path, pathPre***REMOVED***x) {
			if len(pathPre***REMOVED***x) > 1 && len(other.Path) > len(pathPre***REMOVED***x) && other.Path[len(pathPre***REMOVED***x)] != '/' {
				continue
			}
			pathToRoutes.Add(other.Path, other)
		}
	}
	pathToRoutes.Do(func(path string, routes []restful.Route) {
		api := Api{Path: strings.TrimSuf***REMOVED***x(withoutWildcard(path), "/"), Description: ws.Documentation()}
		voidString := "void"
		for _, route := range routes {
			operation := Operation{
				Method:  route.Method,
				Summary: route.Doc,
				Notes:   route.Notes,
				// Type gets overwritten if there is a write sample
				DataTypeFields:   DataTypeFields{Type: &voidString},
				Parameters:       []Parameter{},
				Nickname:         route.Operation,
				ResponseMessages: composeResponseMessages(route, &decl, &sws.con***REMOVED***g)}

			operation.Consumes = route.Consumes
			operation.Produces = route.Produces

			// share root params if any
			for _, swparam := range rootParams {
				operation.Parameters = append(operation.Parameters, swparam)
			}
			// route speci***REMOVED***c params
			for _, param := range route.ParameterDocs {
				operation.Parameters = append(operation.Parameters, asSwaggerParameter(param.Data()))
			}

			sws.addModelsFromRouteTo(&operation, route, &decl)
			api.Operations = append(api.Operations, operation)
		}
		decl.Apis = append(decl.Apis, api)
	})
	return decl
}

func withoutWildcard(path string) string {
	if strings.HasSuf***REMOVED***x(path, ":*}") {
		return path[0:len(path)-3] + "}"
	}
	return path
}

// composeResponseMessages takes the ResponseErrors (if any) and creates ResponseMessages from them.
func composeResponseMessages(route restful.Route, decl *ApiDeclaration, con***REMOVED***g *Con***REMOVED***g) (messages []ResponseMessage) {
	if route.ResponseErrors == nil {
		return messages
	}
	// sort by code
	codes := sort.IntSlice{}
	for code := range route.ResponseErrors {
		codes = append(codes, code)
	}
	codes.Sort()
	for _, code := range codes {
		each := route.ResponseErrors[code]
		message := ResponseMessage{
			Code:    code,
			Message: each.Message,
		}
		if each.Model != nil {
			st := reflect.TypeOf(each.Model)
			isCollection, st := detectCollectionType(st)
			// collection cannot be in responsemodel
			if !isCollection {
				modelName := modelBuilder{}.keyFrom(st)
				modelBuilder{Models: &decl.Models, Con***REMOVED***g: con***REMOVED***g}.addModel(st, "")
				message.ResponseModel = modelName
			}
		}
		messages = append(messages, message)
	}
	return
}

// addModelsFromRoute takes any read or write sample from the Route and creates a Swagger model from it.
func (sws SwaggerService) addModelsFromRouteTo(operation *Operation, route restful.Route, decl *ApiDeclaration) {
	if route.ReadSample != nil {
		sws.addModelFromSampleTo(operation, false, route.ReadSample, &decl.Models)
	}
	if route.WriteSample != nil {
		sws.addModelFromSampleTo(operation, true, route.WriteSample, &decl.Models)
	}
}

func detectCollectionType(st reflect.Type) (bool, reflect.Type) {
	isCollection := false
	if st.Kind() == reflect.Slice || st.Kind() == reflect.Array {
		st = st.Elem()
		isCollection = true
	} ***REMOVED*** {
		if st.Kind() == reflect.Ptr {
			if st.Elem().Kind() == reflect.Slice || st.Elem().Kind() == reflect.Array {
				st = st.Elem().Elem()
				isCollection = true
			}
		}
	}
	return isCollection, st
}

// addModelFromSample creates and adds (or overwrites) a Model from a sample resource
func (sws SwaggerService) addModelFromSampleTo(operation *Operation, isResponse bool, sample interface{}, models *ModelList) {
	mb := modelBuilder{Models: models, Con***REMOVED***g: &sws.con***REMOVED***g}
	if isResponse {
		sampleType, items := asDataType(sample, &sws.con***REMOVED***g)
		operation.Type = sampleType
		operation.Items = items
	}
	mb.addModelFrom(sample)
}

func asSwaggerParameter(param restful.ParameterData) Parameter {
	return Parameter{
		DataTypeFields: DataTypeFields{
			Type:         &param.DataType,
			Format:       asFormat(param.DataType, param.DataFormat),
			DefaultValue: Special(param.DefaultValue),
		},
		Name:        param.Name,
		Description: param.Description,
		ParamType:   asParamType(param.Kind),

		Required: param.Required}
}

// Between 1..7 path parameters is supported
func composeRootPath(req *restful.Request) string {
	path := "/" + req.PathParameter("a")
	b := req.PathParameter("b")
	if b == "" {
		return path
	}
	path = path + "/" + b
	c := req.PathParameter("c")
	if c == "" {
		return path
	}
	path = path + "/" + c
	d := req.PathParameter("d")
	if d == "" {
		return path
	}
	path = path + "/" + d
	e := req.PathParameter("e")
	if e == "" {
		return path
	}
	path = path + "/" + e
	f := req.PathParameter("f")
	if f == "" {
		return path
	}
	path = path + "/" + f
	g := req.PathParameter("g")
	if g == "" {
		return path
	}
	return path + "/" + g
}

func asFormat(dataType string, dataFormat string) string {
	if dataFormat != "" {
		return dataFormat
	}
	return "" // TODO
}

func asParamType(kind int) string {
	switch {
	case kind == restful.PathParameterKind:
		return "path"
	case kind == restful.QueryParameterKind:
		return "query"
	case kind == restful.BodyParameterKind:
		return "body"
	case kind == restful.HeaderParameterKind:
		return "header"
	case kind == restful.FormParameterKind:
		return "form"
	}
	return ""
}

func asDataType(any interface{}, con***REMOVED***g *Con***REMOVED***g) (*string, *Item) {
	// If it's not a collection, return the suggested model name
	st := reflect.TypeOf(any)
	isCollection, st := detectCollectionType(st)
	modelName := modelBuilder{}.keyFrom(st)
	// if it's not a collection we are done
	if !isCollection {
		return &modelName, nil
	}

	// XXX: This is not very elegant
	// We create an Item object referring to the given model
	models := ModelList{}
	mb := modelBuilder{Models: &models, Con***REMOVED***g: con***REMOVED***g}
	mb.addModelFrom(any)

	elemTypeName := mb.getElementTypeName(modelName, "", st)
	item := new(Item)
	if mb.isPrimitiveType(elemTypeName) {
		mapped := mb.jsonSchemaType(elemTypeName)
		item.Type = &mapped
	} ***REMOVED*** {
		item.Ref = &elemTypeName
	}
	tmp := "array"
	return &tmp, item
}
