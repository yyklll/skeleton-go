package api

import (
	"net/http"

	"github.com/yyklll/skeleton/pkg/log"
	"go.uber.org/zap"
)

func (s *Server) reqLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("received request",
			zap.String("proto", r.Proto),
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.String("remote", r.RemoteAddr),
			zap.String("user-agent", r.UserAgent()),
		)
		next.ServeHTTP(w, r)
	})
}
