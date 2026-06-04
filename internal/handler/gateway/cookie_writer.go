package gateway

import (
	"net/http"
	"portal-system/internal/infrastructure/security"
	"strings"
	"time"
)

const (
	accessTokenCookie  = "access_token"
	refreshTokenCookie = "refresh_token"
)

// CookieWritingMiddleware intercepts gRPC-gateway responses and converts
// Grpc-Metadata-Set-Cookie-* and Grpc-Metadata-Clear-Cookie headers into
// actual Set-Cookie headers on the HTTP response.
//
// The gRPC-gateway serializes gRPC response metadata as HTTP response headers
// with the prefix "Grpc-Metadata-". This middleware reads those headers,
// converts them to Set-Cookie, and removes the originals.
func CookieWritingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cw := &cookieResponseWriter{ResponseWriter: w}
		next.ServeHTTP(cw, r)
	})
}

type cookieResponseWriter struct {
	http.ResponseWriter
}

func (cw *cookieResponseWriter) WriteHeader(statusCode int) {
	// Convert Set-Cookie-* metadata headers to actual Set-Cookie headers
	if accessToken := cw.Header().Get("Grpc-Metadata-Set-Cookie-Access-Token"); accessToken != "" {
		http.SetCookie(cw.ResponseWriter, &http.Cookie{
			Name:     accessTokenCookie,
			Value:    accessToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   int((15 * time.Minute).Seconds()),
		})
		cw.Header().Del("Grpc-Metadata-Set-Cookie-Access-Token")
	}

	if refreshToken := cw.Header().Get("Grpc-Metadata-Set-Cookie-Refresh-Token"); refreshToken != "" {
		http.SetCookie(cw.ResponseWriter, &http.Cookie{
			Name:     refreshTokenCookie,
			Value:    refreshToken,
			Path:     "/api/v1/auth",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   int((7 * 24 * time.Hour).Seconds()),
		})
		cw.Header().Del("Grpc-Metadata-Set-Cookie-Refresh-Token")
	}

	if csrfToken := cw.Header().Get("Grpc-Metadata-Set-Cookie-Csrf-Token"); csrfToken != "" {
		http.SetCookie(cw.ResponseWriter, &http.Cookie{
			Name:     security.CSRFCookieName(),
			Value:    csrfToken,
			Path:     "/",
			HttpOnly: false, // JS must read this for Double-Submit
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   int(security.CSRFTTL().Seconds()),
		})
		cw.Header().Del("Grpc-Metadata-Set-Cookie-Csrf-Token")
	}

	// Handle Clear-Cookie metadata (logout)
	if clearValues := cw.Header()["Grpc-Metadata-Clear-Cookie"]; len(clearValues) > 0 {
		for _, name := range clearValues {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			path := "/"
			if name == "refresh_token" {
				path = "/api/v1/auth"
			}
			http.SetCookie(cw.ResponseWriter, &http.Cookie{
				Name:     name,
				Value:    "",
				Path:     path,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
				MaxAge:   -1,
			})
		}
		cw.Header().Del("Grpc-Metadata-Clear-Cookie")
	}

	cw.ResponseWriter.WriteHeader(statusCode)
}
