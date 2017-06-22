// Package sdk is the of***REMOVED***cial AWS SDK for the Go programming language.
//
// The AWS SDK for Go provides APIs and utilities that developers can use to
// build Go applications that use AWS services, such as Amazon Elastic Compute
// Cloud (Amazon EC2) and Amazon Simple Storage Service (Amazon S3).
//
// The SDK removes the complexity of coding directly against a web service
// interface. It hides a lot of the lower-level plumbing, such as authentication,
// request retries, and error handling.
//
// The SDK also includes helpful utilities on top of the AWS APIs that add additional
// capabilities and functionality. For example, the Amazon S3 Download and Upload
// Manager will automatically split up large objects into multiple parts and
// transfer them concurrently.
//
// See the s3manager package documentation for more information.
// https://docs.aws.amazon.com/sdk-for-go/api/service/s3/s3manager/
//
// Getting More Information
//
// Checkout the Getting Started Guide and API Reference Docs detailed the SDK's
// components and details on each AWS client the SDK supports.
//
// The Getting Started Guide provides examples and detailed description of how
// to get setup with the SDK.
// https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/welcome.html
//
// The API Reference Docs include a detailed breakdown of the SDK's components
// such as utilities and AWS clients. Use this as a reference of the Go types
// included with the SDK, such as AWS clients, API operations, and API parameters.
// https://docs.aws.amazon.com/sdk-for-go/api/
//
// Overview of SDK's Packages
//
// The SDK is composed of two main components, SDK core, and service clients.
// The SDK core packages are all available under the aws package at the root of
// the SDK. Each client for a supported AWS service is available within its own
// package under the service folder at the root of the SDK.
//
//   * aws - SDK core, provides common shared types such as Con***REMOVED***g, Logger,
//     and utilities to make working with API parameters easier.
//
//       * awserr - Provides the error interface that the SDK will use for all
//         errors that occur in the SDK's processing. This includes service API
//         response errors as well. The Error type is made up of a code and message.
//         Cast the SDK's returned error type to awserr.Error and call the Code
//         method to compare returned error to speci***REMOVED***c error codes. See the package's
//         documentation for additional values that can be extracted such as RequestId.
//
//       * credentials - Provides the types and built in credentials providers
//         the SDK will use to retrieve AWS credentials to make API requests with.
//         Nested under this folder are also additional credentials providers such as
//         stscreds for assuming IAM roles, and ec2rolecreds for EC2 Instance roles.
//
//       * endpoints - Provides the AWS Regions and Endpoints metadata for the SDK.
//         Use this to lookup AWS service endpoint information such as which services
//         are in a region, and what regions a service is in. Constants are also provided
//         for all region identi***REMOVED***ers, e.g UsWest2RegionID for "us-west-2".
//
//       * session - Provides initial default con***REMOVED***guration, and load
//         con***REMOVED***guration from external sources such as environment and shared
//         credentials ***REMOVED***le.
//
//       * request - Provides the API request sending, and retry logic for the SDK.
//         This package also includes utilities for de***REMOVED***ning your own request
//         retryer, and con***REMOVED***guring how the SDK processes the request.
//
//   * service - Clients for AWS services. All services supported by the SDK are
//     available under this folder.
//
// How to Use the SDK's AWS Service Clients
//
// The SDK includes the Go types and utilities you can use to make requests to
// AWS service APIs. Within the service folder at the root of the SDK you'll ***REMOVED***nd
// a package for each AWS service the SDK supports. All service clients follows
// a common pattern of creation and usage.
//
// When creating a client for an AWS service you'll ***REMOVED***rst need to have a Session
// value constructed. The Session provides shared con***REMOVED***guration that can be shared
// between your service clients. When service clients are created you can pass
// in additional con***REMOVED***guration via the aws.Con***REMOVED***g type to override con***REMOVED***guration
// provided by in the Session to create service client instances with custom
// con***REMOVED***guration.
//
// Once the service's client is created you can use it to make API requests the
// AWS service. These clients are safe to use concurrently.
//
// Con***REMOVED***guring the SDK
//
// In the AWS SDK for Go, you can con***REMOVED***gure settings for service clients, such
// as the log level and maximum number of retries. Most settings are optional;
// however, for each service client, you must specify a region and your credentials.
// The SDK uses these values to send requests to the correct AWS region and sign
// requests with the correct credentials. You can specify these values as part
// of a session or as environment variables.
//
// See the SDK's con***REMOVED***guration guide for more information.
// https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/con***REMOVED***guring-sdk.html
//
// See the session package documentation for more information on how to use Session
// with the SDK.
// https://docs.aws.amazon.com/sdk-for-go/api/aws/session/
//
// See the Con***REMOVED***g type in the aws package for more information on con***REMOVED***guration
// options.
// https://docs.aws.amazon.com/sdk-for-go/api/aws/#Con***REMOVED***g
//
// Con***REMOVED***guring Credentials
//
// When using the SDK you'll generally need your AWS credentials to authenticate
// with AWS services. The SDK supports multiple methods of supporting these
// credentials. By default the SDK will source credentials automatically from
// its default credential chain. See the session package for more information
// on this chain, and how to con***REMOVED***gure it. The common items in the credential
// chain are the following:
//
//   * Environment Credentials - Set of environment variables that are useful
//     when sub processes are created for speci***REMOVED***c roles.
//
//   * Shared Credentials ***REMOVED***le (~/.aws/credentials) - This ***REMOVED***le stores your
//     credentials based on a pro***REMOVED***le name and is useful for local development.
//
//   * EC2 Instance Role Credentials - Use EC2 Instance Role to assign credentials
//     to application running on an EC2 instance. This removes the need to manage
//     credential ***REMOVED***les in production.
//
// Credentials can be con***REMOVED***gured in code as well by setting the Con***REMOVED***g's Credentials
// value to a custom provider or using one of the providers included with the
// SDK to bypass the default credential chain and use a custom one. This is
// helpful when you want to instruct the SDK to only use a speci***REMOVED***c set of
// credentials or providers.
//
// This example creates a credential provider for assuming an IAM role, "myRoleARN"
// and con***REMOVED***gures the S3 service client to use that role for API requests.
//
//   // Initial credentials loaded from SDK's default credential chain. Such as
//   // the environment, shared credentials (~/.aws/credentials), or EC2 Instance
//   // Role. These credentials will be used to to make the STS Assume Role API.
//   sess := session.Must(session.NewSession())
//
//   // Create the credentials from AssumeRoleProvider to assume the role
//   // referenced by the "myRoleARN" ARN.
//   creds := stscreds.NewCredentials(sess, "myRoleArn")
//
//   // Create service client value con***REMOVED***gured for credentials
//   // from assumed role.
//   svc := s3.New(sess, &aws.Con***REMOVED***g{Credentials: creds})/
//
// See the credentials package documentation for more information on credential
// providers included with the SDK, and how to customize the SDK's usage of
// credentials.
// https://docs.aws.amazon.com/sdk-for-go/api/aws/credentials
//
// The SDK has support for the shared con***REMOVED***guration ***REMOVED***le (~/.aws/con***REMOVED***g). This
// support can be enabled by setting the environment variable, "AWS_SDK_LOAD_CONFIG=1",
// or enabling the feature in code when creating a Session via the
// Option's SharedCon***REMOVED***gState parameter.
//
//   sess := session.Must(session.NewSessionWithOptions(session.Options{
//       SharedCon***REMOVED***gState: session.SharedCon***REMOVED***gEnable,
//   }))
//
// Con***REMOVED***guring AWS Region
//
// In addition to the credentials you'll need to specify the region the SDK
// will use to make AWS API requests to. In the SDK you can specify the region
// either with an environment variable, or directly in code when a Session or
// service client is created. The last value speci***REMOVED***ed in code wins if the region
// is speci***REMOVED***ed multiple ways.
//
// To set the region via the environment variable set the "AWS_REGION" to the
// region you want to the SDK to use. Using this method to set the region will
// allow you to run your application in multiple regions without needing additional
// code in the application to select the region.
//
//   AWS_REGION=us-west-2
//
// The endpoints package includes constants for all regions the SDK knows. The
// values are all suf***REMOVED***xed with RegionID. These values are helpful, because they
// reduce the need to type the region string manually.
//
// To set the region on a Session use the aws package's Con***REMOVED***g struct parameter
// Region to the AWS region you want the service clients created from the session to
// use. This is helpful when you want to create multiple service clients, and
// all of the clients make API requests to the same region.
//
//   sess := session.Must(session.NewSession(&aws.Con***REMOVED***g{
//       Region: aws.String(endpoints.UsWest2RegionID),
//   }))
//
// See the endpoints package for the AWS Regions and Endpoints metadata.
// https://docs.aws.amazon.com/sdk-for-go/api/aws/endpoints/
//
// In addition to setting the region when creating a Session you can also set
// the region on a per service client bases. This overrides the region of a
// Session. This is helpful when you want to create service clients in speci***REMOVED***c
// regions different from the Session's region.
//
//   svc := s3.New(sess, &aws.Con***REMOVED***g{
//       Region: aws.String(ednpoints.UsWest2RegionID),
//   })
//
// See the Con***REMOVED***g type in the aws package for more information and additional
// options such as setting the Endpoint, and other service client con***REMOVED***guration options.
// https://docs.aws.amazon.com/sdk-for-go/api/aws/#Con***REMOVED***g
//
// Making API Requests
//
// Once the client is created you can make an API request to the service.
// Each API method takes a input parameter, and returns the service response
// and an error. The SDK provides methods for making the API call in multiple ways.
//
// In this list we'll use the S3 ListObjects API as an example for the different
// ways of making API requests.
//
//   * ListObjects - Base API operation that will make the API request to the service.
//
//   * ListObjectsRequest - API methods suf***REMOVED***xed with Request will construct the
//     API request, but not send it. This is also helpful when you want to get a
//     presigned URL for a request, and share the presigned URL instead of your
//     application making the request directly.
//
//   * ListObjectsPages - Same as the base API operation, but uses a callback to
//     automatically handle pagination of the API's response.
//
//   * ListObjectsWithContext - Same as base API operation, but adds support for
//     the Context pattern. This is helpful for controlling the canceling of in
//     flight requests. See the Go standard library context package for more
//     information. This method also takes request package's Option functional
//     options as the variadic argument for modifying how the request will be
//     made, or extracting information from the raw HTTP response.
//
//   * ListObjectsPagesWithContext - same as ListObjectsPages, but adds support for
//     the Context pattern. Similar to ListObjectsWithContext this method also
//     takes the request package's Option function option types as the variadic
//     argument.
//
// In addition to the API operations the SDK also includes several higher level
// methods that abstract checking for and waiting for an AWS resource to be in
// a desired state. In this list we'll use WaitUntilBucketExists to demonstrate
// the different forms of waiters.
//
//   * WaitUntilBucketExists. - Method to make API request to query an AWS service for
//     a resource's state. Will return successfully when that state is accomplished.
//
//   * WaitUntilBucketExistsWithContext - Same as WaitUntilBucketExists, but adds
//     support for the Context pattern. In addition these methods take request
//     package's WaiterOptions to con***REMOVED***gure the waiter, and how underlying request
//     will be made by the SDK.
//
// The API method will document which error codes the service might return for
// the operation. These errors will also be available as const strings pre***REMOVED***xed
// with "ErrCode" in the service client's package. If there are no errors listed
// in the API's SDK documentation you'll need to consult the AWS service's API
// documentation for the errors that could be returned.
//
//   ctx := context.Background()
//
//   result, err := svc.GetObjectWithContext(ctx, &s3.GetObjectInput{
//       Bucket: aws.String("my-bucket"),
//       Key: aws.String("my-key"),
//   })
//   if err != nil {
//       // Cast err to awserr.Error to handle speci***REMOVED***c error codes.
//       aerr, ok := err.(awserr.Error)
//       if ok && aerr.Code() == s3.ErrCodeNoSuchKey {
//           // Speci***REMOVED***c error code handling
//       }
//       return err
//   }
//
//   // Make sure to close the body when done with it for S3 GetObject APIs or
//   // will leak connections.
//   defer result.Body.Close()
//
//   fmt.Println("Object Size:", aws.StringValue(result.ContentLength))
//
// API Request Pagination and Resource Waiters
//
// Pagination helper methods are suf***REMOVED***xed with "Pages", and provide the
// functionality needed to round trip API page requests. Pagination methods
// take a callback function that will be called for each page of the API's response.
//
//    objects := []string{}
//    err := svc.ListObjectsPagesWithContext(ctx, &s3.ListObjectsInput{
//        Bucket: aws.String(myBucket),
//    }, func(p *s3.ListObjectsOutput, lastPage bool) bool {
//        for _, o := range p.Contents {
//            objects = append(objects, aws.StringValue(o.Key))
//        }
//        return true // continue paging
//    })
//    if err != nil {
//        panic(fmt.Sprintf("failed to list objects for bucket, %s, %v", myBucket, err))
//    }
//
//    fmt.Println("Objects in bucket:", objects)
//
// Waiter helper methods provide the functionality to wait for an AWS resource
// state. These methods abstract the logic needed to to check the state of an
// AWS resource, and wait until that resource is in a desired state. The waiter
// will block until the resource is in the state that is desired, an error occurs,
// or the waiter times out. If a resource times out the error code returned will
// be request.WaiterResourceNotReadyErrorCode.
//
//   err := svc.WaitUntilBucketExistsWithContext(ctx, &s3.HeadBucketInput{
//       Bucket: aws.String(myBucket),
//   })
//   if err != nil {
//       aerr, ok := err.(awserr.Error)
//       if ok && aerr.Code() == request.WaiterResourceNotReadyErrorCode {
//           fmt.Fprintf(os.Stderr, "timed out while waiting for bucket to exist")
//       }
//       panic(fmt.Errorf("failed to wait for bucket to exist, %v", err))
//   }
//   fmt.Println("Bucket", myBucket, "exists")
//
// Complete SDK Example
//
// This example shows a complete working Go ***REMOVED***le which will upload a ***REMOVED***le to S3
// and use the Context pattern to implement timeout logic that will cancel the
// request if it takes too long. This example highlights how to use sessions,
// create a service client, make a request, handle the error, and process the
// response.
//
//   package main
//
//   import (
//   	"context"
//   	"flag"
//   	"fmt"
//   	"os"
//   	"time"
//
//   	"github.com/aws/aws-sdk-go/aws"
//   	"github.com/aws/aws-sdk-go/aws/awserr"
//   	"github.com/aws/aws-sdk-go/aws/request"
//   	"github.com/aws/aws-sdk-go/aws/session"
//   	"github.com/aws/aws-sdk-go/service/s3"
//   )
//
//   // Uploads a ***REMOVED***le to S3 given a bucket and object key. Also takes a duration
//   // value to terminate the update if it doesn't complete within that time.
//   //
//   // The AWS Region needs to be provided in the AWS shared con***REMOVED***g or on the
//   // environment variable as `AWS_REGION`. Credentials also must be provided
//   // Will default to shared con***REMOVED***g ***REMOVED***le, but can load from environment if provided.
//   //
//   // Usage:
//   //   # Upload my***REMOVED***le.txt to myBucket/myKey. Must complete within 10 minutes or will fail
//   //   go run withContext.go -b mybucket -k myKey -d 10m < my***REMOVED***le.txt
//   func main() {
//   	var bucket, key string
//   	var timeout time.Duration
//
//   	flag.StringVar(&bucket, "b", "", "Bucket name.")
//   	flag.StringVar(&key, "k", "", "Object key name.")
//   	flag.DurationVar(&timeout, "d", 0, "Upload timeout.")
//   	flag.Parse()
//
//   	// All clients require a Session. The Session provides the client with
//  	// shared con***REMOVED***guration such as region, endpoint, and credentials. A
//  	// Session should be shared where possible to take advantage of
//  	// con***REMOVED***guration and credential caching. See the session package for
//  	// more information.
//   	sess := session.Must(session.NewSession())
//
//  	// Create a new instance of the service's client with a Session.
//  	// Optional aws.Con***REMOVED***g values can also be provided as variadic arguments
//  	// to the New function. This option allows you to provide service
//  	// speci***REMOVED***c con***REMOVED***guration.
//   	svc := s3.New(sess)
//
//   	// Create a context with a timeout that will abort the upload if it takes
//   	// more than the passed in timeout.
//   	ctx := context.Background()
//   	var cancelFn func()
//   	if timeout > 0 {
//   		ctx, cancelFn = context.WithTimeout(ctx, timeout)
//   	}
//   	// Ensure the context is canceled to prevent leaking.
//   	// See context package for more information, https://golang.org/pkg/context/
//   	defer cancelFn()
//
//   	// Uploads the object to S3. The Context will interrupt the request if the
//   	// timeout expires.
//   	_, err := svc.PutObjectWithContext(ctx, &s3.PutObjectInput{
//   		Bucket: aws.String(bucket),
//   		Key:    aws.String(key),
//   		Body:   os.Stdin,
//   	})
//   	if err != nil {
//   		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
//   			// If the SDK can determine the request or retry delay was canceled
//   			// by a context the CanceledErrorCode error code will be returned.
//   			fmt.Fprintf(os.Stderr, "upload canceled due to timeout, %v\n", err)
//   		} ***REMOVED*** {
//   			fmt.Fprintf(os.Stderr, "failed to upload object, %v\n", err)
//   		}
//   		os.Exit(1)
//   	}
//
//   	fmt.Printf("successfully uploaded ***REMOVED***le to %s/%s\n", bucket, key)
//   }
package sdk
