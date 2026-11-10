package main

import (
	"net/http"
	"time"
)

const (
	idCookieName      = "music-id"
	refreshCookieName = "music-refresh"
)

// setAuthCookie sets one of the auth cookies. In production the cookies are
// secure, http only, and scoped to the configured cookie domain so the
// frontend and api subdomains can share them.
func (app *application) setAuthCookie(w http.ResponseWriter, name, value string, expires time.Time) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Expires:  expires,
		SameSite: http.SameSiteLaxMode,
	}

	if app.config.env == "production" {
		cookie.Domain = app.config.cookieDomain
		cookie.Secure = true
		cookie.HttpOnly = true
		cookie.SameSite = http.SameSiteNoneMode
	}

	http.SetCookie(w, cookie)
}

// clearAuthCookies expires the id and refresh cookies
func (app *application) clearAuthCookies(w http.ResponseWriter) {
	expired := time.Now().Add(-1 * time.Hour)
	app.setAuthCookie(w, idCookieName, "", expired)
	app.setAuthCookie(w, refreshCookieName, "", expired)
}
