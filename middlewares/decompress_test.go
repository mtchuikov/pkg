package middlewares

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zlib"
	"github.com/stretchr/testify/suite"
)

type testDecompressSuite struct {
	suite.Suite

	payload     []byte
	payloadSize int64
	body        *bytes.Buffer

	req        *http.Request
	rr         *httptest.ResponseRecorder
	middleware http.Handler
}

func TestDecompressSuite(t *testing.T) {
	suite.Run(t, new(testDecompressSuite))
}

func (s *testDecompressSuite) SetupTest() {
	s.payload = []byte("hello world")
	s.payloadSize = int64(len(s.payload))
	s.body = &bytes.Buffer{}

	s.req = httptest.NewRequest(http.MethodGet, "/", nil)
	s.rr = httptest.NewRecorder()

	handler := func(rw http.ResponseWriter, req *http.Request) {
		payload, err := io.ReadAll(req.Body)

		errMsg := "request body must be read"
		s.Require().NoError(err, errMsg)

		defer req.Body.Close()

		errMsg = "invalid payload"
		s.Require().Equal(s.payload, payload, errMsg)

		errMsg = "invalid content length"
		s.Require().Equal(s.payloadSize, req.ContentLength)
	}

	s.middleware = Decompress(http.HandlerFunc(handler))
}

func (s *testDecompressSuite) TestDecompress_NoCompression() {
	s.req.Body = io.NopCloser(bytes.NewBuffer(s.payload))
	s.req.ContentLength = s.payloadSize

	s.middleware.ServeHTTP(s.rr, s.req)
}

func (s *testDecompressSuite) TestDecompress_Deflate() {
	s.payloadSize = -1

	zwr := zlib.NewWriter(s.body)
	zwr.Write(s.payload)
	zwr.Close()

	s.req.Body = io.NopCloser(s.body)
	s.req.ContentLength = s.payloadSize
	s.req.Header.Set("Content-Encoding", "deflate")

	s.middleware.ServeHTTP(s.rr, s.req)

	errMsg := "status code must be 200"
	s.Require().Equal(http.StatusOK, s.rr.Code, errMsg)
}

func (s *testDecompressSuite) TestDecompress_Gzip() {
	s.payloadSize = -1

	gwr := gzip.NewWriter(s.body)
	gwr.Write(s.payload)
	gwr.Close()

	s.req.Body = io.NopCloser(s.body)
	s.req.ContentLength = s.payloadSize
	s.req.Header.Set("Content-Encoding", "gzip")

	s.middleware.ServeHTTP(s.rr, s.req)

	errMsg := "status code must be 200"
	s.Require().Equal(http.StatusOK, s.rr.Code, errMsg)
}

func (s *testDecompressSuite) TestDecompress_DeflateBadPayload() {
	s.req.Body = io.NopCloser(bytes.NewBuffer(s.payload))
	s.req.Header.Set("Content-Encoding", "deflate")

	s.middleware.ServeHTTP(s.rr, s.req)

	errMsg := "status code must be 400"
	s.Require().Equal(http.StatusBadRequest, s.rr.Code, errMsg)
}

func (s *testDecompressSuite) TestDecompress_GzipBadPayload() {
	s.req.Body = io.NopCloser(bytes.NewBuffer(s.payload))
	s.req.Header.Set("Content-Encoding", "gzip")

	s.middleware.ServeHTTP(s.rr, s.req)

	errMsg := "status code must be 400"
	s.Require().Equal(http.StatusBadRequest, s.rr.Code, errMsg)
}
