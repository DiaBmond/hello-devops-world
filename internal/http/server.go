package http

import (
	"log/slog"
	"net/http"
	"time"

	"go-prod-app/internal/service"
)

func StartServer(
	userService *service.UserService,
	logger *slog.Logger,
) *http.Server {

	mux := http.NewServeMux()

	handler := NewHandler(userService)
	RegisterRoutes(mux, handler)

	var h http.Handler = mux
	h = MetricsMiddleware()(h)
	h = RecoveryMiddleware(logger)(h)
	h = RequestIDMiddleware(logger)(h)
	h = TimeoutMiddleware(10 * time.Second)(h)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      h,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("http server started", "port", 8080)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
		}
	}()

	return server
}
