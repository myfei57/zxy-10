package console

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"edgelog/internal/engine"
)

// Server exposes the EdgeLog control plane over HTTP: four console pages and
// the JSON API used by the pages and by integration tooling.
type Server struct {
	engine *engine.Engine
	router chi.Router
}

// NewServer creates the console server and registers all routes.
func NewServer(e *engine.Engine) *Server {
	s := &Server{engine: e}
	s.router = chi.NewRouter()
	s.routes()
	return s
}

// Handler returns the HTTP handler for the server.
func (s *Server) Handler() http.Handler {
	return s.router
}

// ListenAndServe starts the HTTP server on addr.
func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.Handler())
}
