package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSetAuthCookie(t *testing.T) {
	tests := []struct {
		env          string
		wantDomain   string
		wantSecure   bool
		wantHTTPOnly bool
		wantSameSite http.SameSite
	}{
		{env: "development", wantSameSite: http.SameSiteLaxMode},
		{env: "production", wantDomain: "example.com", wantSecure: true, wantHTTPOnly: true, wantSameSite: http.SameSiteNoneMode},
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			app := &application{config: config{env: tt.env, cookieDomain: "example.com"}}
			rr := httptest.NewRecorder()

			app.setAuthCookie(rr, idCookieName, "token", time.Now().Add(time.Hour))

			cookies := rr.Result().Cookies()
			if len(cookies) != 1 {
				t.Fatalf("want 1 cookie, got %d", len(cookies))
			}
			c := cookies[0]
			if c.Name != idCookieName || c.Value != "token" || c.Path != "/" {
				t.Errorf("unexpected cookie %+v", c)
			}
			if c.Domain != tt.wantDomain {
				t.Errorf("domain: want %q, got %q", tt.wantDomain, c.Domain)
			}
			if c.Secure != tt.wantSecure || c.HttpOnly != tt.wantHTTPOnly || c.SameSite != tt.wantSameSite {
				t.Errorf("want secure=%v httpOnly=%v sameSite=%v, got %+v", tt.wantSecure, tt.wantHTTPOnly, tt.wantSameSite, c)
			}
		})
	}
}

func TestClearAuthCookies(t *testing.T) {
	app := &application{config: config{env: "development"}}
	rr := httptest.NewRecorder()

	app.clearAuthCookies(rr)

	cookies := rr.Result().Cookies()
	if len(cookies) != 2 {
		t.Fatalf("want 2 cookies, got %d", len(cookies))
	}
	for _, c := range cookies {
		if c.Value != "" || c.Expires.After(time.Now()) {
			t.Errorf("cookie %s was not cleared: %+v", c.Name, c)
		}
	}
}

func TestEnvOr(t *testing.T) {
	t.Setenv("MUSIC_TUTOR_TEST_VAR", "")
	if got := envOr("MUSIC_TUTOR_TEST_VAR", "fallback"); got != "fallback" {
		t.Errorf("want fallback, got %q", got)
	}
	t.Setenv("MUSIC_TUTOR_TEST_VAR", "set")
	if got := envOr("MUSIC_TUTOR_TEST_VAR", "fallback"); got != "set" {
		t.Errorf("want set, got %q", got)
	}
}
