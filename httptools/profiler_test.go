package httptools

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegisterProfilerHandlers(t *testing.T) {
	mux := http.NewServeMux()
	RegisterProfilerHandlers(mux)

	paths := []string{
		"/debug/pprof/",
		"/debug/pprof/cmdline",
		// Don't want to test accessibility of this two commented endpoints
		// accessability cause their invocation triggers long-running tasks

		// "/debug/pprof/profile",
		"/debug/pprof/symbol",
		// "/debug/pprof/trace",
		"/debug/pprof/allocs",
		"/debug/pprof/vars",
		"/debug/pprof/goroutine",
		"/debug/pprof/heap",
		"/debug/pprof/threadcreate",
		"/debug/pprof/mutex",
		"/debug/pprof/block",
	}

	for _, path := range paths {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		errMsg := "expected 200 status code when calling %s"
		require.Equalf(t, http.StatusOK, rr.Code, errMsg, path)
	}
}
