package middlewares

import (
	"net/http"
	"strings"
)

// AllowContentTypes is an HTTP middleware that restricts
// allowed Content-Types. It takes a variadic list of permitted
// content types and ensures the request's Content-Type matches
// one (case-insensitive). The Content-Type is extracted,
// parameters are removed, and it's compared to the whitelist.
// If not allowed, it returns a 415 Unsupported Media Type
// error. Otherwise, it passes to the next handler. Note that
// Content-Type values are normalized to lowercase and trimmed,
// and empty Content-Type is invalid unless explicitly allowed.
func AllowContentTypes(contentTypes ...string) func(http.Handler) http.Handler {
	contentTypesLen := len(contentTypes)
	whitelist := make(map[string]struct{}, contentTypesLen)

	for _, contentType := range contentTypes {
		contentType = strings.TrimSpace(strings.ToLower(contentType))
		whitelist[contentType] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		fn := func(rw http.ResponseWriter, req *http.Request) {
			contentType := req.Header.Get("Content-Type")
			if contentType != "" {
				contentType = strings.Split(contentType, ";")[0]
				contentType = strings.TrimSpace(strings.ToLower(contentType))
			}

			_, whitelisted := whitelist[contentType]
			if !whitelisted {
				errMsg := "invalid content type"
				http.Error(rw, errMsg, http.StatusUnsupportedMediaType)
				return
			}

			next.ServeHTTP(rw, req)
		}

		return http.HandlerFunc(fn)
	}
}
