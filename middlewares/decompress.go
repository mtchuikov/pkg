package middlewares

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"io"
	"net/http"
	"sync"
)

var gzipReaderPool = sync.Pool{
	New: func() any {
		return &gzip.Reader{}
	},
}

type pooledGzipReader struct {
	gr   *gzip.Reader
	pool *sync.Pool
}

func (r *pooledGzipReader) Read(b []byte) (int, error) {
	return r.gr.Read(b)
}

func (r *pooledGzipReader) Close() error {
	err := r.gr.Close()
	r.pool.Put(r.gr)
	return err
}

var zlibInitReaderBytes = []byte{0x78, 0x9c, 0x01, 0x00, 0x00, 0xff, 0xff, 0x00, 0x00, 0x00, 0x01}

var zlibReaderPool = sync.Pool{
	New: func() any {
		br := bytes.NewReader(zlibInitReaderBytes)
		zr, _ := zlib.NewReader(br)
		return zr
	},
}

type pooledZlibReader struct {
	zr   io.ReadCloser
	pool *sync.Pool
}

func (r *pooledZlibReader) Read(p []byte) (int, error) {
	return r.zr.Read(p)
}

func (r *pooledZlibReader) Close() error {
	err := r.zr.Close()
	r.pool.Put(r.zr)
	return err
}

func resetContentEncoding(req *http.Request) {
	req.Header.Del("Content-Encoding")
	req.ContentLength = -1
}

// Decompress is an HTTP middleware that decompresses request
// body. It supports gzip and deflate encoded requests only as
// most popular one. Before calling this middleware, ensure
// that a request encoding format are verified for support. If
// not, the request body, encoded with, for example, Brotli,
// will be passed directly to the handler, which will not be
// able to properly read it.
func Decompress(next http.Handler) http.Handler {
	fn := func(rw http.ResponseWriter, req *http.Request) {
		encoding := req.Header.Get("Content-Encoding")
		switch encoding {
		case "deflate":
			zr := zlibReaderPool.Get().(io.ReadCloser)
			err := zr.(zlib.Resetter).Reset(req.Body, nil)
			if err != nil {
				errMsg := "failed to decompress deflated body"
				http.Error(rw, errMsg, http.StatusBadRequest)
				return
			}
			defer zr.Close()

			resetContentEncoding(req)
			req.Body = &pooledZlibReader{
				zr:   zr,
				pool: &zlibReaderPool,
			}

		case "gzip":
			gr := gzipReaderPool.Get().(*gzip.Reader)
			err := gr.Reset(req.Body)
			if err != nil {
				errMsg := "failed to decompress gzipped body"
				http.Error(rw, errMsg, http.StatusBadRequest)
				return
			}
			defer gr.Close()

			resetContentEncoding(req)
			req.Body = &pooledGzipReader{
				gr:   gr,
				pool: &gzipReaderPool,
			}
		}

		next.ServeHTTP(rw, req)
	}

	return http.HandlerFunc(fn)
}
