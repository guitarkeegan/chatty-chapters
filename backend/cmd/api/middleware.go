package main

import (
	"context"
	"fmt"
	"net/http"
)

type contextKey string

const (
	userCognitoSubKey contextKey = "cognitoSub"
	TEACHER           string     = "teacher"
	ADMIN             string     = "admin"
	STUDENT           string     = "student"
)

// TODO: don't allow from any origin
func (app *application) enableCoors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		w.Header().Add("Vary", "Access-Control-Request-Method")
		w.Header().Add("Vary", "Access-Control-Request-Headers")

		origin := r.Header.Get("Origin")

		if origin != "" {
			for i := range app.config.cors.trustedOrigins {
				if origin == app.config.cors.trustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Credentials", "true")

					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {

						w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

						w.WriteHeader(http.StatusOK)
						return
					}

					break
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dbg("start authenticate")

		musicID, err := r.Cookie("music-id")
		if err != nil {
			app.invalidAuthenticationTokenResponse(w, r, err)
			return
		}

		claims, err := app.verifyToken(musicID.Value)
		if err != nil {
			// Token is invalid or expired, try to refresh
			refreshToken, err := r.Cookie("music-refresh")
			if err != nil {
				app.invalidAuthenticationTokenResponse(w, r, err)
				return
			}

			refreshResp, err := app.refreshToken(refreshToken.Value)
			if err != nil {
				app.invalidAuthenticationTokenResponse(w, r, err)
				return
			}

			// Update the cookies with new tokens
			app.sendIDAndRefreshToClient(w, http.StatusOK, envelope{"status": "token_refreshed"}, refreshResp.AuthenticationResult)

			// Update claims with new token
			claims, err = app.verifyToken(*refreshResp.AuthenticationResult.IdToken)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}
		}

		ctx := context.WithValue(r.Context(), userCognitoSubKey, claims.Subject)
		dbg("got sub: %s", claims.Subject)
		dbg("calling next")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) authenticateTeacher(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		role, err := app.getRole(r)
		if err != nil {
			app.serverErrorResponse(w, r, fmt.Errorf("authenticateTeacher: %w", err))
			return
		}

		if role != TEACHER {
			app.invalidAuthenticationTokenResponse(w, r, fmt.Errorf("authenicateTeacher: %w", err))
			return
		}

		next.ServeHTTP(w, r)

	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		app.logger.Info("received", "route", r.URL.RequestURI(), "method", r.Method)

		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
