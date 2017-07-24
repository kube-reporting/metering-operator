package swagger

type SwaggerBuilder struct {
	SwaggerService
}

func NewSwaggerBuilder(con***REMOVED***g Con***REMOVED***g) *SwaggerBuilder {
	return &SwaggerBuilder{*newSwaggerService(con***REMOVED***g)}
}

func (sb SwaggerBuilder) ProduceListing() ResourceListing {
	return sb.SwaggerService.produceListing()
}

func (sb SwaggerBuilder) ProduceAllDeclarations() map[string]ApiDeclaration {
	return sb.SwaggerService.produceAllDeclarations()
}

func (sb SwaggerBuilder) ProduceDeclarations(route string) (*ApiDeclaration, bool) {
	return sb.SwaggerService.produceDeclarations(route)
}
