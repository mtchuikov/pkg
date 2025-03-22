package httptools

import (
	"expvar"
	"net/http"
	"net/http/pprof"
)

const ProfilerHandlersPrefix = "/debug/pprof"

// SetupProfilerHandlers registers HTTP handlers for profiling
// using the net/http/pprof package, setting up endpoints under
// /debug/pprof/ for accessing profiling data like CPU, memory,
// and goroutine profiles etc.
func RegisterProfilerHandlers(mux *http.ServeMux) {
	mux.HandleFunc(ProfilerHandlersPrefix+"/", pprof.Index)

	mux.HandleFunc(ProfilerHandlersPrefix+"/cmdline", pprof.Cmdline)
	mux.HandleFunc(ProfilerHandlersPrefix+"/profile", pprof.Profile)
	mux.HandleFunc(ProfilerHandlersPrefix+"/symbol", pprof.Symbol)
	mux.HandleFunc(ProfilerHandlersPrefix+"/trace", pprof.Trace)

	mux.Handle(ProfilerHandlersPrefix+"/allocs", pprof.Handler("allocs"))
	mux.Handle(ProfilerHandlersPrefix+"/vars", expvar.Handler())

	mux.Handle(ProfilerHandlersPrefix+"/goroutine", pprof.Handler("goroutine"))
	mux.Handle(ProfilerHandlersPrefix+"/heap", pprof.Handler("heap"))
	mux.Handle(ProfilerHandlersPrefix+"/threadcreate", pprof.Handler("threadcreate"))
	mux.Handle(ProfilerHandlersPrefix+"/mutex", pprof.Handler("mutex"))
	mux.Handle(ProfilerHandlersPrefix+"/block", pprof.Handler("block"))
}
