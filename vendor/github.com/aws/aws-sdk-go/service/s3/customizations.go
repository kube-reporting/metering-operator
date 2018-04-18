package s3

import (
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/request"
)

func init() {
	initClient = defaultInitClientFn
	initRequest = defaultInitRequestFn
}

func defaultInitClientFn(c *client.Client) {
	// Support building custom endpoints based on con***REMOVED***g
	c.Handlers.Build.PushFront(updateEndpointForS3Con***REMOVED***g)

	// Require SSL when using SSE keys
	c.Handlers.Validate.PushBack(validateSSERequiresSSL)
	c.Handlers.Build.PushBack(computeSSEKeys)

	// S3 uses custom error unmarshaling logic
	c.Handlers.UnmarshalError.Clear()
	c.Handlers.UnmarshalError.PushBack(unmarshalError)
}

func defaultInitRequestFn(r *request.Request) {
	// Add reuest handlers for speci***REMOVED***c platforms.
	// e.g. 100-continue support for PUT requests using Go 1.6
	platformRequestHandlers(r)

	switch r.Operation.Name {
	case opPutBucketCors, opPutBucketLifecycle, opPutBucketPolicy,
		opPutBucketTagging, opDeleteObjects, opPutBucketLifecycleCon***REMOVED***guration,
		opPutBucketReplication:
		// These S3 operations require Content-MD5 to be set
		r.Handlers.Build.PushBack(contentMD5)
	case opGetBucketLocation:
		// GetBucketLocation has custom parsing logic
		r.Handlers.Unmarshal.PushFront(buildGetBucketLocation)
	case opCreateBucket:
		// Auto-populate LocationConstraint with current region
		r.Handlers.Validate.PushFront(populateLocationConstraint)
	case opCopyObject, opUploadPartCopy, opCompleteMultipartUpload:
		r.Handlers.Unmarshal.PushFront(copyMultipartStatusOKUnmarhsalError)
	case opPutObject, opUploadPart:
		r.Handlers.Build.PushBack(computeBodyHashes)
		// Disabled until #1837 root issue is resolved.
		//	case opGetObject:
		//		r.Handlers.Build.PushBack(askForTxEncodingAppendMD5)
		//		r.Handlers.Unmarshal.PushBack(useMD5ValidationReader)
	}
}

// bucketGetter is an accessor interface to grab the "Bucket" ***REMOVED***eld from
// an S3 type.
type bucketGetter interface {
	getBucket() string
}

// sseCustomerKeyGetter is an accessor interface to grab the "SSECustomerKey"
// ***REMOVED***eld from an S3 type.
type sseCustomerKeyGetter interface {
	getSSECustomerKey() string
}

// copySourceSSECustomerKeyGetter is an accessor interface to grab the
// "CopySourceSSECustomerKey" ***REMOVED***eld from an S3 type.
type copySourceSSECustomerKeyGetter interface {
	getCopySourceSSECustomerKey() string
}
