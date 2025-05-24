package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"shortening-api/internal/database"
	"shortening-api/internal/helpers"
	"time"
)

const (
	RefreshTokenValidDays = 7
)

type loginForm struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (app *application) loginHandler(w http.ResponseWriter, r *http.Request) {
	var lgnForm loginForm
	if err := helpers.ParseForm(r, &lgnForm); err != nil {
		app.serverError(w, r, err)
		return
	}

	user, err := app.queries.GetUser(r.Context(), lgnForm.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helpers.InvalidCredentialsResponse(w)
			return
		}
		app.serverError(w, r, err)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(lgnForm.Password)); err != nil {
		helpers.InvalidCredentialsResponse(w)
		return
	}

	expiration := time.Now().Add(time.Hour * 24 * RefreshTokenValidDays)
	refreshToken, err := createToken(user.ID.String(), expiration)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  expiration,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	aToken, err := createToken(user.ID.String(), time.Now().Add(time.Minute*15))
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	type Token struct {
		AccessToken string `json:"access_token"`
	}
	accessToken := Token{AccessToken: aToken}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&accessToken)

	if err != nil {
		app.serverError(w, r, err)
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			Value:    "",
			MaxAge:   -1,
		})
	}
}

type SignUpForm struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (app *application) signUpHandler(w http.ResponseWriter, r *http.Request) {
	var signUpForm SignUpForm
	if err := helpers.ParseForm(r, &signUpForm); err != nil {
		app.badRequest(w, r, err)
		return
	}

	if Matches(signUpForm.Email, EmailRX) || !Blank(signUpForm.Password) || MinChars(signUpForm.Password, 8) {
		app.badRequest(w, r, fmt.Errorf("invalid signing up credentials"))
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(signUpForm.Password), bcrypt.DefaultCost+2)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	user, err := app.queries.CreateUser(r.Context(), database.CreateUserParams{
		ID:           uuid.New(),
		Email:        signUpForm.Email,
		PasswordHash: string(passwordHash),
	})
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	expiration := time.Now().Add(time.Hour * 24 * RefreshTokenValidDays)
	refreshToken, err := createToken(user.ID.String(), expiration)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  expiration,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	aToken, err := createToken(user.ID.String(), time.Now().Add(time.Minute*15))
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	type Token struct {
		AccessToken string `json:"access_token"`
	}
	accessToken := Token{AccessToken: aToken}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(&accessToken)

	if err != nil {
		app.serverError(w, r, err)
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			Value:    "",
			MaxAge:   -1,
		})
	}
}

func (app *application) logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			app.unauthorized(w, r, err)
			return
		}
		app.badRequest(w, r, err)
		return
	}
	tokenString := cookie.Value

	token, err := verifyToken(tokenString)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	jtiString := token.Claims.(*jwt.RegisteredClaims).ID
	userIDString, _ := token.Claims.GetSubject()
	expiresAt := token.Claims.(*jwt.RegisteredClaims).ExpiresAt

	jti, err := uuid.Parse(jtiString)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	userID, err := uuid.Parse(userIDString)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	revoked, err := app.queries.IsTokenRevoked(r.Context(), jti)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if revoked {
		app.badRequest(w, r, err)
		return
	}

	_, err = app.queries.RevokeJwt(r.Context(), database.RevokeJwtParams{
		Jti:       jti,
		UserID:    userID,
		ExpiresAt: expiresAt.Time,
	})
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	app.logger.Debug("Token revoked", "jti", jti.String(), " serID:", userID, "expiresAt:", expiresAt.String())

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		Value:    "",
		MaxAge:   -1,
	})
}

func (app *application) refreshHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			app.unauthorized(w, r, err)
			return
		}
		app.badRequest(w, r, err)
		return
	}
	tokenString := cookie.Value

	token, err := verifyToken(tokenString)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	jtiString := token.Claims.(*jwt.RegisteredClaims).ID
	userIDString, _ := token.Claims.GetSubject()
	expiresAt := token.Claims.(*jwt.RegisteredClaims).ExpiresAt

	jti, err := uuid.Parse(jtiString)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	userID, err := uuid.Parse(userIDString)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	revoked, err := app.queries.IsTokenRevoked(r.Context(), jti)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if revoked {
		app.badRequest(w, r, err)
		return
	}

	_, err = app.queries.RevokeJwt(r.Context(), database.RevokeJwtParams{
		Jti:       jti,
		UserID:    userID,
		ExpiresAt: expiresAt.Time,
	})
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	user, err := app.queries.GetUserByID(r.Context(), userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	expiration := time.Now().Add(time.Hour * 24 * RefreshTokenValidDays)
	refreshToken, err := createToken(user.ID.String(), expiration)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  expiration,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	aToken, err := createToken(user.ID.String(), time.Now().Add(time.Minute*15))
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	type Token struct {
		AccessToken string `json:"access_token"`
	}
	accessToken := Token{AccessToken: aToken}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&accessToken)

	if err != nil {
		app.serverError(w, r, err)
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			Value:    "",
			MaxAge:   -1,
		})
	}
}

func (app *application) pubKeyHandler(w http.ResponseWriter, r *http.Request) {
	file, err := os.ReadFile("config/keys/public_key.pem")
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/x-pem-file")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(file)
}
