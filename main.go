package main

import (
	"cmp"
	"crypto/sha256"
	"log/slog"
	"main/internal/config"
	"main/internal/logger"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware(sha256.Sum256([]byte(config.Global.AuthToken)), "auth_token"))

		r.Get("/check_replicas", checkReplicasHandler(config.Global.TargetURL, httpClient))
	})

	// Health endpoint outside auth group
	r.Get("/health", healthHandler)

	port := cmp.Or(os.Getenv("PORT"), "8080")

	logger.Stdout.Info("starting server", slog.String("port", port))

	if err := http.ListenAndServe((":" + port), r); err != nil {
		logger.Stderr.Error("failed to start server", logger.ErrAttr(err))
	}
}
