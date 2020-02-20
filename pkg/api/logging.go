package api

import (
	"net/http"

	"github.com/yyklll/skeleton/pkg/log"
)

func (s *Server) reqLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("received request",
			"proto", r.Proto,
			"uri", r.RequestURI,
			"method", r.Method,
			"remote", r.RemoteAddr,
			"user-agent", r.UserAgent(),
		)
		next.ServeHTTP(w, r)
	})
}
