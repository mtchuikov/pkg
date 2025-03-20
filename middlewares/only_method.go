package middlewares

import "net/http"

// OnlyMethod is an HTTP middleware that restricts request handling
// to a specific HTTP method. If a request's method does not match
// the specified method, it responds with a 405 Method Not Allowed
// status. This middleware ensures that only requests with the
// allowed method are processed by the subsequent handler.
func OnlyMethod(method string, next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if method != req.Method {
			errMsg := "method not allowed"
			http.Error(rw, errMsg, http.StatusMethodNotAllowed)
			return
		}

		next(rw, req)
	}
}
