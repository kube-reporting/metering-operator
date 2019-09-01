// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE ***REMOVED***le.

package appengine

import (
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/internal"
	pb "google.golang.org/appengine/internal/app_identity"
	modpb "google.golang.org/appengine/internal/modules"
)

// AppID returns the application ID for the current application.
// The string will be a plain application ID (e.g. "appid"), with a
// domain pre***REMOVED***x for custom domain deployments (e.g. "example.com:appid").
func AppID(c context.Context) string { return internal.AppID(c) }

// DefaultVersionHostname returns the standard hostname of the default version
// of the current application (e.g. "my-app.appspot.com"). This is suitable for
// use in constructing URLs.
func DefaultVersionHostname(c context.Context) string {
	return internal.DefaultVersionHostname(c)
}

// ModuleName returns the module name of the current instance.
func ModuleName(c context.Context) string {
	return internal.ModuleName(c)
}

// ModuleHostname returns a hostname of a module instance.
// If module is the empty string, it refers to the module of the current instance.
// If version is empty, it refers to the version of the current instance if valid,
// or the default version of the module of the current instance.
// If instance is empty, ModuleHostname returns the load-balancing hostname.
func ModuleHostname(c context.Context, module, version, instance string) (string, error) {
	req := &modpb.GetHostnameRequest{}
	if module != "" {
		req.Module = &module
	}
	if version != "" {
		req.Version = &version
	}
	if instance != "" {
		req.Instance = &instance
	}
	res := &modpb.GetHostnameResponse{}
	if err := internal.Call(c, "modules", "GetHostname", req, res); err != nil {
		return "", err
	}
	return *res.Hostname, nil
}

// VersionID returns the version ID for the current application.
// It will be of the form "X.Y", where X is speci***REMOVED***ed in app.yaml,
// and Y is a number generated when each version of the app is uploaded.
// It does not include a module name.
func VersionID(c context.Context) string { return internal.VersionID(c) }

// InstanceID returns a mostly-unique identi***REMOVED***er for this instance.
func InstanceID() string { return internal.InstanceID() }

// Datacenter returns an identi***REMOVED***er for the datacenter that the instance is running in.
func Datacenter(c context.Context) string { return internal.Datacenter(c) }

// ServerSoftware returns the App Engine release version.
// In production, it looks like "Google App Engine/X.Y.Z".
// In the development appserver, it looks like "Development/X.Y".
func ServerSoftware() string { return internal.ServerSoftware() }

// RequestID returns a string that uniquely identi***REMOVED***es the request.
func RequestID(c context.Context) string { return internal.RequestID(c) }

// AccessToken generates an OAuth2 access token for the speci***REMOVED***ed scopes on
// behalf of service account of this application. This token will expire after
// the returned time.
func AccessToken(c context.Context, scopes ...string) (token string, expiry time.Time, err error) {
	req := &pb.GetAccessTokenRequest{Scope: scopes}
	res := &pb.GetAccessTokenResponse{}

	err = internal.Call(c, "app_identity_service", "GetAccessToken", req, res)
	if err != nil {
		return "", time.Time{}, err
	}
	return res.GetAccessToken(), time.Unix(res.GetExpirationTime(), 0), nil
}

// Certi***REMOVED***cate represents a public certi***REMOVED***cate for the app.
type Certi***REMOVED***cate struct {
	KeyName string
	Data    []byte // PEM-encoded X.509 certi***REMOVED***cate
}

// PublicCerti***REMOVED***cates retrieves the public certi***REMOVED***cates for the app.
// They can be used to verify a signature returned by SignBytes.
func PublicCerti***REMOVED***cates(c context.Context) ([]Certi***REMOVED***cate, error) {
	req := &pb.GetPublicCerti***REMOVED***cateForAppRequest{}
	res := &pb.GetPublicCerti***REMOVED***cateForAppResponse{}
	if err := internal.Call(c, "app_identity_service", "GetPublicCerti***REMOVED***catesForApp", req, res); err != nil {
		return nil, err
	}
	var cs []Certi***REMOVED***cate
	for _, pc := range res.PublicCerti***REMOVED***cateList {
		cs = append(cs, Certi***REMOVED***cate{
			KeyName: pc.GetKeyName(),
			Data:    []byte(pc.GetX509Certi***REMOVED***catePem()),
		})
	}
	return cs, nil
}

// ServiceAccount returns a string representing the service account name, in
// the form of an email address (typically app_id@appspot.gserviceaccount.com).
func ServiceAccount(c context.Context) (string, error) {
	req := &pb.GetServiceAccountNameRequest{}
	res := &pb.GetServiceAccountNameResponse{}

	err := internal.Call(c, "app_identity_service", "GetServiceAccountName", req, res)
	if err != nil {
		return "", err
	}
	return res.GetServiceAccountName(), err
}

// SignBytes signs bytes using a private key unique to your application.
func SignBytes(c context.Context, bytes []byte) (keyName string, signature []byte, err error) {
	req := &pb.SignForAppRequest{BytesToSign: bytes}
	res := &pb.SignForAppResponse{}

	if err := internal.Call(c, "app_identity_service", "SignForApp", req, res); err != nil {
		return "", nil, err
	}
	return res.GetKeyName(), res.GetSignatureBytes(), nil
}

func init() {
	internal.RegisterErrorCodeMap("app_identity_service", pb.AppIdentityServiceError_ErrorCode_name)
	internal.RegisterErrorCodeMap("modules", modpb.ModulesServiceError_ErrorCode_name)
}
