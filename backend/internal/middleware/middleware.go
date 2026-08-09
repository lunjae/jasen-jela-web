package middleware

import (
	"context"
	"github.com/lunjae/jasen-jela-web/backend/internal/response"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

type key string

const uidKey key = "uid"

type TokenVerifier interface {
	Verify(context.Context, string) (string, error)
}
type AdminChecker interface {
	IsAdmin(context.Context, string) (bool, error)
}
type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, m ...Middleware) http.Handler {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}
func RequestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start))
	})
}
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if x := recover(); x != nil {
				slog.Error("panic recovered", "error", x, "stack", string(debug.Stack()))
				response.WriteError(w, &response.APIError{Status: 500, Code: "internal_error", Message: "Došlo je do greške."})
			}
		}()
		next.ServeHTTP(w, r)
	})
}
func CORS(origin string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Origin") == origin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Dev-Admin-UID")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(204)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
func Authenticate(v TokenVerifier, development bool) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			uid := ""
			if development {
				uid = r.Header.Get("X-Dev-Admin-UID")
			}
			if uid == "" {
				parts := strings.Fields(r.Header.Get("Authorization"))
				if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
					response.WriteError(w, &response.APIError{Status: 401, Code: "unauthorized", Message: "Prijavite se za nastavak."})
					return
				}
				var err error
				uid, err = v.Verify(r.Context(), parts[1])
				if err != nil {
					response.WriteError(w, &response.APIError{Status: 401, Code: "unauthorized", Message: "Sesija nije važeća."})
					return
				}
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), uidKey, uid)))
		})
	}
}
func RequireAdmin(c AdminChecker) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			uid, _ := r.Context().Value(uidKey).(string)
			ok, err := c.IsAdmin(r.Context(), uid)
			if err != nil || !ok {
				response.WriteError(w, &response.APIError{Status: 403, Code: "forbidden", Message: "Nemate administratorski pristup."})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
