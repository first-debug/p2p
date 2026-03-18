package websocket

import (
	"log/slog"
	"net/http"
)

func (s *WebSocketServer) stopingMiddleware(fn func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.isStopping.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			if _, err := w.Write([]byte("server is stopping")); err != nil {
				s.logger.Error("cannot emit about server status", slog.String("error", err.Error()))
			}
			return
		}
		fn(w, r)
	}
}
