package gophercloud

import (
	"fmt"
	"net/url"
	"path/***REMOVED***lepath"
	"strings"
	"time"
)

// WaitFor polls a predicate function, once per second, up to a timeout limit.
// This is useful to wait for a resource to transition to a certain state.
// To handle situations when the predicate might hang inde***REMOVED***nitely, the
// predicate will be prematurely cancelled after the timeout.
// Resource packages will wrap this in a more convenient function that's
// speci***REMOVED***c to a certain resource, but it can also be useful on its own.
func WaitFor(timeout int, predicate func() (bool, error)) error {
	type WaitForResult struct {
		Success bool
		Error   error
	}

	start := time.Now().Unix()

	for {
		// If a timeout is set, and that's been exceeded, shut it down.
		if timeout >= 0 && time.Now().Unix()-start >= int64(timeout) {
			return fmt.Errorf("A timeout occurred")
		}

		time.Sleep(1 * time.Second)

		var result WaitForResult
		ch := make(chan bool, 1)
		go func() {
			defer close(ch)
			satis***REMOVED***ed, err := predicate()
			result.Success = satis***REMOVED***ed
			result.Error = err
		}()

		select {
		case <-ch:
			if result.Error != nil {
				return result.Error
			}
			if result.Success {
				return nil
			}
		// If the predicate has not ***REMOVED***nished by the timeout, cancel it.
		case <-time.After(time.Duration(timeout) * time.Second):
			return fmt.Errorf("A timeout occurred")
		}
	}
}

// NormalizeURL is an internal function to be used by provider clients.
//
// It ensures that each endpoint URL has a closing `/`, as expected by
// ServiceClient's methods.
func NormalizeURL(url string) string {
	if !strings.HasSuf***REMOVED***x(url, "/") {
		return url + "/"
	}
	return url
}

// NormalizePathURL is used to convert rawPath to a fqdn, using basePath as
// a reference in the ***REMOVED***lesystem, if necessary. basePath is assumed to contain
// either '.' when ***REMOVED***rst used, or the ***REMOVED***le:// type fqdn of the parent resource.
// e.g. myFavScript.yaml => ***REMOVED***le://opt/lib/myFavScript.yaml
func NormalizePathURL(basePath, rawPath string) (string, error) {
	u, err := url.Parse(rawPath)
	if err != nil {
		return "", err
	}
	// if a scheme is de***REMOVED***ned, it must be a fqdn already
	if u.Scheme != "" {
		return u.String(), nil
	}
	// if basePath is a url, then child resources are assumed to be relative to it
	bu, err := url.Parse(basePath)
	if err != nil {
		return "", err
	}
	var basePathSys, absPathSys string
	if bu.Scheme != "" {
		basePathSys = ***REMOVED***lepath.FromSlash(bu.Path)
		absPathSys = ***REMOVED***lepath.Join(basePathSys, rawPath)
		bu.Path = ***REMOVED***lepath.ToSlash(absPathSys)
		return bu.String(), nil
	}

	absPathSys = ***REMOVED***lepath.Join(basePath, rawPath)
	u.Path = ***REMOVED***lepath.ToSlash(absPathSys)
	if err != nil {
		return "", err
	}
	u.Scheme = "***REMOVED***le"
	return u.String(), nil

}
