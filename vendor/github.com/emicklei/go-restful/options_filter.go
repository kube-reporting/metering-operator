package restful

import "strings"

// Copyright 2013 Ernest Micklei. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE ***REMOVED***le.

// OPTIONSFilter is a ***REMOVED***lter function that inspects the Http Request for the OPTIONS method
// and provides the response with a set of allowed methods for the request URL Path.
// As for any ***REMOVED***lter, you can also install it for a particular WebService within a Container.
// Note: this ***REMOVED***lter is not needed when using CrossOriginResourceSharing (for CORS).
func (c *Container) OPTIONSFilter(req *Request, resp *Response, chain *FilterChain) {
	if "OPTIONS" != req.Request.Method {
		chain.ProcessFilter(req, resp)
		return
	}
	resp.AddHeader(HEADER_Allow, strings.Join(c.computeAllowedMethods(req), ","))
}

// OPTIONSFilter is a ***REMOVED***lter function that inspects the Http Request for the OPTIONS method
// and provides the response with a set of allowed methods for the request URL Path.
// Note: this ***REMOVED***lter is not needed when using CrossOriginResourceSharing (for CORS).
func OPTIONSFilter() FilterFunction {
	return DefaultContainer.OPTIONSFilter
}
