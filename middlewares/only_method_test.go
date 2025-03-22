package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func testOnlyMethod(whitelistedMethod string) (int, bool) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	var handlerCalled bool
	handler := func(rw http.ResponseWriter, req *http.Request) {
		handlerCalled = true
	}

	rr := httptest.NewRecorder()
	middleware := OnlyMethod(whitelistedMethod, http.HandlerFunc(handler))
	middleware.ServeHTTP(rr, req)

	return rr.Code, handlerCalled
}

func TestOnlyMethod_Allowed(t *testing.T) {
	code, handlerCalled := testOnlyMethod(http.MethodGet)

	errMsg := "expected status code 200, got %v"
	require.Equalf(t, http.StatusOK, code, errMsg, code)

	require.True(t, handlerCalled, "expected handler to be called")
}

func TestOnlyMethod_NotAllowed(t *testing.T) {
	code, handlerCalled := testOnlyMethod(http.MethodPost)

	errMsg := "expected status code 405, got %v"
	require.Equalf(t, http.StatusMethodNotAllowed, code, errMsg, code)

	require.False(t, handlerCalled, "expected handler not to be called")
}
