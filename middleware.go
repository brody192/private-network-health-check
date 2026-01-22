package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
)

func authMiddleware(expectedTokenHash [32]byte, queryKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			providedTokenHash := sha256.Sum256([]byte(r.URL.Query().Get(queryKey)))
			if subtle.ConstantTimeCompare(providedTokenHash[:], expectedTokenHash[:]) != 1 {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
