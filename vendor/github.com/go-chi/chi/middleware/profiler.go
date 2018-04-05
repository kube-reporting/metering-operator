package middleware

import (
	"expvar"
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi"
)

// Pro***REMOVED***ler is a convenient subrouter used for mounting net/http/pprof. ie.
//
//  func MyService() http.Handler {
//    r := chi.NewRouter()
//    // ..middlewares
//    r.Mount("/debug", middleware.Pro***REMOVED***ler())
//    // ..routes
//    return r
//  }
func Pro***REMOVED***ler() http.Handler {
	r := chi.NewRouter()
	r.Use(NoCache)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.RequestURI+"/pprof/", 301)
	})
	r.HandleFunc("/pprof", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.RequestURI+"/", 301)
	})

	r.HandleFunc("/pprof/*", pprof.Index)
	r.HandleFunc("/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/pprof/pro***REMOVED***le", pprof.Pro***REMOVED***le)
	r.HandleFunc("/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/pprof/trace", pprof.Trace)
	r.HandleFunc("/vars", expVars)

	return r
}

// Replicated from expvar.go as not public.
func expVars(w http.ResponseWriter, r *http.Request) {
	***REMOVED***rst := true
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\n")
	expvar.Do(func(kv expvar.KeyValue) {
		if !***REMOVED***rst {
			fmt.Fprintf(w, ",\n")
		}
		***REMOVED***rst = false
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(w, "\n}\n")
}
