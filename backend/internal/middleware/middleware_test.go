package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type verifier struct {
	uid string
	err error
}

func (v verifier) Verify(context.Context, string) (string, error) { return v.uid, v.err }

type checker bool

func (c checker) IsAdmin(context.Context, string) (bool, error) { return bool(c), nil }
func TestAuthenticationRejectsInvalidToken(t *testing.T) {
	h := Authenticate(verifier{err: errors.New("bad")}, false)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer invalid")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatalf("status=%d", w.Code)
	}
}
func TestAuthenticationRejectsMissingTokenBeforeImageHandler(t *testing.T) {
	called := false
	h := Authenticate(verifier{uid: "admin"}, false)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))
	r := httptest.NewRequest(http.MethodPost, "/api/admin/products/p/images", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized || called {
		t.Fatalf("status=%d called=%v", w.Code, called)
	}
}
func TestAdminRoleComesFromStore(t *testing.T) {
	called := false
	end := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })
	h := Chain(end, Authenticate(verifier{uid: "user"}, false), RequireAdmin(checker(false)))
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer token")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 403 || called {
		t.Fatalf("status=%d called=%v", w.Code, called)
	}
}
