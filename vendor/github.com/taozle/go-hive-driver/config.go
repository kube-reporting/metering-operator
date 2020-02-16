package hive

import (
	"errors"
	"net/url"
	"strconv"
	"time"
)

var ErrNoPassword = errors.New("hive: password is required")

type ConnectOptions struct {
	Host      string
	Timeout   time.Duration
	AuthMode  string
	Username  string
	Password  string
	BatchSize int64
}

func connectOptionsFromURL(u *url.URL) (ConnectOptions, error) {
	var opts ConnectOptions

	opts.Host = u.Host
	queryParams := u.Query()

	opts.AuthMode = queryParams.Get("auth")

	if batchSizeParam := queryParams.Get("batch"); batchSizeParam != "" {
		var err error
		if opts.BatchSize, err = strconv.ParseInt(batchSizeParam, 10, 64); err != nil {
			return ConnectOptions{}, err
		}
	}

	if timeoutParam := queryParams.Get("connect_timeout"); timeoutParam != "" {
		if timeoutSeconds, err := strconv.ParseInt(timeoutParam, 10, 0); err != nil {
			return ConnectOptions{}, err
		} else {
			opts.Timeout = time.Duration(timeoutSeconds) * time.Second
		}
	}

	switch opts.AuthMode {
	case "sasl":
		if name := u.User.Username(); name != "" {
			opts.Username = name
		}
		if password, ok := u.User.Password(); ok {
			opts.Password = password
		} else {
			return ConnectOptions{}, ErrNoPassword
		}
	}

	return opts, nil
}
