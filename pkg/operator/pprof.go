package operator

import (
	"net/http"
	"net/http/pprof"
)

func newPprofServer() *http.Server {
	pprofMux := http.NewServeMux()

	pprofMux.HandleFunc("/debug/pprof/", pprof.Index)
	pprofMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	pprofMux.HandleFunc("/debug/pprof/pro***REMOVED***le", pprof.Pro***REMOVED***le)
	pprofMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	pprofMux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return &http.Server{
		Addr:    "127.0.0.1:6060",
		Handler: pprofMux,
	}
}
