package console

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) routes() {
	s.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ui/pipelines", http.StatusFound)
	})
	s.router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	s.router.Get("/ui/pipelines", s.page("pipelines.html"))
	s.router.Get("/ui/config", s.page("config.html"))
	s.router.Get("/ui/rules", s.page("rules.html"))
	s.router.Get("/ui/deadletter", s.page("deadletter.html"))

	s.router.Route("/api", func(r chi.Router) {
		r.Get("/status", s.status)
		r.Route("/sources", func(r chi.Router) {
			r.Get("/", s.listSources)
			r.Post("/", s.createSource)
			r.Post("/{id}/ingest", s.ingestSource)
		})
		r.Route("/sinks", func(r chi.Router) {
			r.Get("/", s.listSinks)
			r.Post("/", s.createSink)
			r.Post("/{id}/fail", s.failSink)
		})
		r.Route("/pipelines", func(r chi.Router) {
			r.Get("/", s.listPipelines)
			r.Post("/", s.switchPipeline)
		})
		r.Route("/rules", func(r chi.Router) {
			r.Get("/", s.getRules)
			r.Post("/", s.updateRules)
		})
		r.Route("/deadletters", func(r chi.Router) {
			r.Get("/", s.listDeadLetters)
			r.Post("/{id}/requeue", s.requeueDeadLetter)
		})
		r.Get("/audit", s.auditList)
	})
}
