package protocol

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
)

// HostPre***REMOVED***xHandlerName is the handler name for the host pre***REMOVED***x request
// handler.
const HostPre***REMOVED***xHandlerName = "awssdk.endpoint.HostPre***REMOVED***xHandler"

// NewHostPre***REMOVED***xHandler constructs a build handler
func NewHostPre***REMOVED***xHandler(pre***REMOVED***x string, labelsFn func() map[string]string) request.NamedHandler {
	builder := HostPre***REMOVED***xBuilder{
		Pre***REMOVED***x:   pre***REMOVED***x,
		LabelsFn: labelsFn,
	}

	return request.NamedHandler{
		Name: HostPre***REMOVED***xHandlerName,
		Fn:   builder.Build,
	}
}

// HostPre***REMOVED***xBuilder provides the request handler to expand and prepend
// the host pre***REMOVED***x into the operation's request endpoint host.
type HostPre***REMOVED***xBuilder struct {
	Pre***REMOVED***x   string
	LabelsFn func() map[string]string
}

// Build updates the passed in Request with the HostPre***REMOVED***x template expanded.
func (h HostPre***REMOVED***xBuilder) Build(r *request.Request) {
	if aws.BoolValue(r.Con***REMOVED***g.DisableEndpointHostPre***REMOVED***x) {
		return
	}

	var labels map[string]string
	if h.LabelsFn != nil {
		labels = h.LabelsFn()
	}

	pre***REMOVED***x := h.Pre***REMOVED***x
	for name, value := range labels {
		pre***REMOVED***x = strings.Replace(pre***REMOVED***x, "{"+name+"}", value, -1)
	}

	r.HTTPRequest.URL.Host = pre***REMOVED***x + r.HTTPRequest.URL.Host
	if len(r.HTTPRequest.Host) > 0 {
		r.HTTPRequest.Host = pre***REMOVED***x + r.HTTPRequest.Host
	}
}
