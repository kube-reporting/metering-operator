package aws

import "github.com/aws/aws-sdk-go/aws/awserr"

var (
	// ErrMissingRegion is an error that is returned if region con***REMOVED***guration is
	// not found.
	//
	// @readonly
	ErrMissingRegion = awserr.New("MissingRegion", "could not ***REMOVED***nd region con***REMOVED***guration", nil)

	// ErrMissingEndpoint is an error that is returned if an endpoint cannot be
	// resolved for a service.
	//
	// @readonly
	ErrMissingEndpoint = awserr.New("MissingEndpoint", "'Endpoint' con***REMOVED***guration is required for this service", nil)
)
