package httptools

import (
	"fmt"
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
		// don't want to test accessibility of this two
		// commented endpoints accessability cause their
		// invocation triggers long-running tasks

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

		errMsg := fmt.Sprintf("expected status code 200, got %v", rr.Code)
		require.Equal(t, http.StatusOK, rr.Code, errMsg)
	}
}
