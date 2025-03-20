package middlewares

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestVerbose(t *testing.T) {
	const (
		mockMethod                = http.MethodGet
		mockURL                   = "/test"
		mockUserAgent             = "test-agent"
		mockReferer               = "http://example.com"
		mockRemoteAddr            = "127.0.0.1:12345"
		mockResponseStatus        = http.StatusOK
		mockResponseBody          = "Hello, World!"
		mockResponseContentLenght = 13
	)

	req := httptest.NewRequest(mockMethod, mockURL, nil)
	req.Header.Set("User-Agent", mockUserAgent)
	req.Header.Set("Referer", mockReferer)
	req.RemoteAddr = mockRemoteAddr

	handler := func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(mockResponseStatus)
		rw.Write([]byte(mockResponseBody))
	}

	var logBuf bytes.Buffer
	logger := zerolog.New(&logBuf).With().
		Timestamp().Logger()

	verbose := Verbose(logger)
	middleware := verbose(http.HandlerFunc(handler))

	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	output := logBuf.String()
	assert.Contains(t, output, fmt.Sprintf(`"method":"%v"`, mockMethod))
	assert.Contains(t, output, fmt.Sprintf(`"url":"%v"`, mockURL))
	assert.Contains(t, output, fmt.Sprintf(`"remote_addr":"%v`, mockRemoteAddr))
	assert.Contains(t, output, fmt.Sprintf(`"user_agent":"%v"`, mockUserAgent))
	assert.Contains(t, output, fmt.Sprintf(`"status":%v`, mockResponseStatus))
	assert.Contains(t, output, fmt.Sprintf(`"size":%v`, mockResponseContentLenght))
	assert.Contains(t, output, fmt.Sprintf(`"referer":"%v"`, mockReferer))
	assert.Contains(t, output, `"duration":`)
}
