package console

import (
	"embed"
	"net/http"
)

//go:embed webui/*.html
var pagesFS embed.FS

// page returns a handler that serves one embedded console page.
func (s *Server) page(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := pagesFS.ReadFile("webui/" + name)
		if err != nil {
			http.Error(w, "page not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(data)
	}
}
