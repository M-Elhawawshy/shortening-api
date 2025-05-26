package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"log"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"shortening-api/internal/database"
	"shortening-api/internal/helpers"
	"strings"
)

type application struct {
	logger  *slog.Logger
	queries *database.Queries
}

func main() {
	db, err := helpers.OpenDB()
	if err != nil {
		log.Fatal(err)
	}
	port, err := helpers.GetEnv("GATEWAY_PORT")
	if err != nil {
		log.Fatal(err)
	}
	authPort, err := helpers.GetEnv("AUTH_PORT")
	if err != nil {
		log.Fatal(err)
	}
	redirectPort, err := helpers.GetEnv("REDIRECT_PORT")
	if err != nil {
		log.Fatal(err)
	}
	shortenerPort, err := helpers.GetEnv("SHORTENER_PORT")
	if err != nil {
		log.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))

	queries := database.New(db)

	app := application{
		logger:  logger,
		queries: queries,
	}
	r := chi.NewRouter()
	r.Use(app.logRequest, app.recoverPanic)
	r.Route("/api", func(r chi.Router) {
		r.Mount("/auth", app.proxyHandler("http://localhost:"+authPort))
		// use auth middleware
		r.Route("/shorten", func(r chi.Router) {
			r.Use(app.authMiddleware)
			r.Handle("/*", app.proxyHandler("http://localhost:"+shortenerPort))
		})
		r.Route("/redirect", func(r chi.Router) {
			r.Use(app.authMiddleware)
			r.Handle("/*", app.proxyHandler("http://localhost:"+redirectPort))
		})
	})
	log.Println("Auth app is listening on port: " + port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

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

func (app *application) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// todo: unimplemented

		next.ServeHTTP(w, r)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			ip     = r.RemoteAddr
			proto  = r.Proto
			method = r.Method
			uri    = r.URL.RequestURI()
		)

		app.logger.Info("received request", "ip", ip, "proto", proto, "method", method, "uri", uri)

		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")

				app.serverError(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Error(err.Error(), "method: ", r.Method, " uri: ", r.RequestURI)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, r *http.Request, err error, status int) {
	app.logger.Error(err.Error(), "method: ", r.Method, " uri: ", r.RequestURI)
	http.Error(w, http.StatusText(status), status)
}
