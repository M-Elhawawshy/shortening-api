package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang-jwt/jwt/v5/request"
	"io"
	"log"
	"net/http"
	"shortening-api/internal/helpers"
)

func (app *application) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gatewayPort, err := helpers.GetEnv("GATEWAY_PORT")
		if err != nil {
			log.Fatal(err)
		}

		url := "http://localhost:" + gatewayPort + "/api/auth/public.pem"
		resp, err := http.Get(url)
		if err != nil {
			app.serverError(w, r, fmt.Errorf("failed to fetch public key: %w", err))
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			app.serverError(w, r, fmt.Errorf("could not get pub key from request"))
			app.logger.Debug("status code: " + resp.Status)
			return
		}
		publicKeyPem, err := io.ReadAll(resp.Body)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
		app.logger.Debug("we have the pub key pem")
		block, _ := pem.Decode(publicKeyPem)
		if block == nil {
			app.serverError(w, r, fmt.Errorf("failed to decode pub key"))
			return
		}
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
		verifyKey, ok := pub.(*rsa.PublicKey)
		if !ok {
			app.serverError(w, r, err)
			return
		}
		token, err := request.ParseFromRequest(r, request.OAuth2Extractor, func(token *jwt.Token) (interface{}, error) {
			// use public key to verify token
			return verifyKey, nil
		}, request.WithClaims(&jwt.RegisteredClaims{}))
		if err != nil {
			app.clientError(w, r, err, http.StatusUnauthorized)
			return
		}
		// putting the user id in the context
		userID := token.Claims.(*jwt.RegisteredClaims).Subject
		ctx := context.WithValue(r.Context(), helpers.UserIDKey, userID)
		r = r.WithContext(ctx)

		app.logger.Debug("userID: " + userID)
		app.logger.Debug("access_token: " + token.Raw)
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
