package apiserver

import "net/http"

func (s *Server) onlineMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.online {
			http.Error(w, "API server offline", http.StatusServiceUnavailable)
			return
		}
		next.ServeHTTP(w, r)
	})
}
