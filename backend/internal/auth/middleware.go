package auth

import (
	"context"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
)

// ParentOnly middleware verifies the JWT and ensures role == "parent".
// Returns 403 Forbidden if the token has a child role.
// (Assumes jwtauth.Verifier and jwtauth.Authenticator have already run.)
func ParentOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := GetRoleFromContext(r.Context())
		if role != "parent" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ChildOnly middleware verifies the JWT and ensures role == "child".
// Returns 403 Forbidden if the token has a parent role.
// (Assumes jwtauth.Verifier and jwtauth.Authenticator have already run.)
func ChildOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := GetRoleFromContext(r.Context())
		if role != "child" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// GetUserIDFromContext extracts the "sub" claim from the JWT stored in ctx.
func GetUserIDFromContext(ctx context.Context) string {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil || claims == nil {
		return ""
	}
	sub, _ := claims["sub"].(string)
	return sub
}

// GetRoleFromContext extracts the "role" claim from the JWT stored in ctx.
func GetRoleFromContext(ctx context.Context) string {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil || claims == nil {
		return ""
	}
	role, _ := claims["role"].(string)
	return role
}
