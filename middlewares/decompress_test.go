package middlewares

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecompress_NoCompression(t *testing.T) {
	const mockPayload = "payload"
	const mockContentLength = int64(len(mockPayload))

	body := strings.NewReader(mockPayload)
	req := httptest.NewRequest(http.MethodGet, "/", body)
	req.ContentLength = mockContentLength

	handler := func(rw http.ResponseWriter, req *http.Request) {
		body, _ := io.ReadAll(req.Body)
		payload := string(body)

		errMsg := "expected body '%s', got %s"
		require.Equalf(t, mockPayload, payload, errMsg, mockPayload, payload)

		errMsg = "expected content lenght to be %v, got %v"
		require.Equalf(t, mockContentLength, req.ContentLength, errMsg,
			mockContentLength, req.ContentLength)
	}

	rr := httptest.NewRecorder()
	middleware := Decompress(http.HandlerFunc(handler))
	middleware.ServeHTTP(rr, req)

	errMsg := "expected status code 200, got %v"
	require.Equalf(t, http.StatusOK, rr.Code, errMsg, rr.Code)
}

func testDecompress_Success(
	t *testing.T,
	w io.WriteCloser, body *bytes.Buffer,
	contentEncoding string,
) {
	const mockPayload = "payload"
	const mockContentLength = int64(-1)

	w.Write([]byte(mockPayload))
	w.Close()

	req := httptest.NewRequest(http.MethodGet, "/", body)
	req.Header.Set("Content-Encoding", contentEncoding)

	handler := func(rw http.ResponseWriter, req *http.Request) {
		body, _ := io.ReadAll(req.Body)
		payload := string(body)

		errMsg := "expected body '%s', got '%s'"
		require.Equalf(t, "payload", mockPayload, errMsg, mockPayload, payload)

		contentEncoding := req.Header.Get("Content-Encoding")
		errMsg = "expected content encoding to be removed, got %s"
		require.Emptyf(t, contentEncoding, errMsg, contentEncoding)

		errMsg = "expected content lenght to be %v, got %v"
		require.Equalf(t, mockContentLength, req.ContentLength, errMsg,
			mockContentLength, req.ContentLength)
	}

	rr := httptest.NewRecorder()
	middleware := Decompress(http.HandlerFunc(handler))
	middleware.ServeHTTP(rr, req)

	errMsg := fmt.Sprintf("expected status code 200, got %v", rr.Code)
	require.Equal(t, http.StatusOK, rr.Code, errMsg)
}

func TestDecompress_Deflate(t *testing.T) {
	var body bytes.Buffer
	zl := zlib.NewWriter(&body)
	testDecompress_Success(t, zl, &body, "deflate")
}

func TestDecompress_Gzip(t *testing.T) {
	var body bytes.Buffer
	gz := gzip.NewWriter(&body)
	testDecompress_Success(t, gz, &body, "gzip")
}

func testDecompress_InvalidBody(
	t *testing.T,
	contentEncoding string,
) {
	const mockPayload = "payload"

	body := strings.NewReader(mockPayload)
	req := httptest.NewRequest(http.MethodGet, "/", body)
	req.Header.Set("Content-Encoding", contentEncoding)

	var handlerCalled bool
	handler := func(rw http.ResponseWriter, req *http.Request) {
		handlerCalled = true
	}

	rr := httptest.NewRecorder()
	middleware := Decompress(http.HandlerFunc(handler))
	middleware.ServeHTTP(rr, req)

	errMsg := "handler shouldn't be called on invalid body"
	require.False(t, handlerCalled, errMsg)

	errMsg = "expected status code 400, got %v"
	require.Equalf(t, http.StatusBadRequest, rr.Code, errMsg, rr.Code)
}

func TestDecompress_InvalidDeflate(t *testing.T) {
	testDecompress_InvalidBody(t, "deflate")
}

func TestDecompress_InvalidGzip(t *testing.T) {
	testDecompress_InvalidBody(t, "gzip")
}
