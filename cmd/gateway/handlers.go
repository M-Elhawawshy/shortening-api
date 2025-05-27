package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httputil"
	"net/url"
	"shortening-api/internal/helpers"
	"strings"
)

func (app *application) proxyHandler(serviceBaseURL string) http.Handler {
	targetURL, err := url.Parse(serviceBaseURL)
	if err != nil {
		panic("invalid proxy target: " + err.Error())
	}
	app.logger.Debug("Proxy target URL: " + targetURL.String())

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			routePattern := chi.RouteContext(req.Context()).RoutePattern()

			prefix := routePattern
			if strings.HasSuffix(prefix, "/*") {
				prefix = prefix[:len(prefix)-2]
			}

			suffix := strings.TrimPrefix(req.URL.Path, prefix)
			if suffix == "" {
				suffix = "/"
			}

			app.logger.Debug("Route pattern: " + routePattern)
			app.logger.Debug("Prefix used for trimming: " + prefix)
			app.logger.Debug("Suffix path: " + suffix)

			req.URL.Scheme = targetURL.Scheme
			req.URL.Host = targetURL.Host
			req.URL.Path = singleJoiningSlash(targetURL.Path, suffix)
			req.URL.RawPath = ""
			req.Host = targetURL.Host
			ctx := req.Context()
			if userID, ok := ctx.Value(helpers.UserIDKey).(string); ok {
				app.logger.Debug("forwarding user ID into http headers")
				req.Header.Set("X-User-ID", userID)
			}
			app.logger.Debug("Proxied URL path: " + req.URL.Path)
		},
	}
	return proxy
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:] // remove duplicate slash
	case !aslash && !bslash:
		return a + "/" + b // add missing slash
	}
	return a + b // one slash already present
}
