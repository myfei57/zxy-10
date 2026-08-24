package console

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/go-chi/chi/v5"

	"edgelog/internal/config"
)

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func decodeJSON(r *http.Request, value any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(value)
}

func (s *Server) status(w http.ResponseWriter, r *http.Request) {
	sources, err := s.engine.Tailer.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	batchSummary, err := s.engine.Batches.Summary()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	sinkSummary, err := s.engine.Sinks.Summary()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	ruleset, err := s.engine.Rules.Current()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	topics, err := s.engine.Batches.TopicSummary()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	deadLetters, err := s.engine.DeadLetters.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	view := make([]map[string]any, 0, len(sources))
	for _, source := range sources {
		buf := s.engine.Router.Buffer(source.ID)
		stats := map[string]any{}
		if buf != nil {
			snapshot := buf.Snapshot()
			stats = map[string]any{"capacity": snapshot.Capacity, "size": snapshot.Size, "dropped": snapshot.Dropped}
		}
		pipeline, active := s.engine.Pipelines.Active(source.ID)
		view = append(view, map[string]any{
			"source":    source,
			"buffer":    stats,
			"pipeline":  pipeline,
			"active":    active,
			"throttled": s.engine.Flow.IsThrottled(source.ID),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"pipelines":   view,
		"batches":     batchSummary,
		"sinks":       sinkSummary,
		"rules":       ruleset,
		"deadletters": len(deadLetters),
		"topics":      topics,
	})
}

func (s *Server) listSources(w http.ResponseWriter, r *http.Request) {
	all, err := s.engine.Tailer.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, all)
}

func (s *Server) createSource(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.Path == "" {
		req.Path = filepath.Join(s.engine.CfgDataDir(), "source-"+strconv.FormatInt(nowNanos(), 10)+".log")
	}
	source, err := s.engine.RegisterSource(req.Name, req.Path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, source)
}

func (s *Server) ingestSource(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Lines []string `json:"lines"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.engine.Ingest(chi.URLParam(r, "id"), req.Lines); err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ingested"})
}

func (s *Server) listSinks(w http.ResponseWriter, r *http.Request) {
	all, err := s.engine.Sinks.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, all)
}

func (s *Server) createSink(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.engine.Sinks.Register(req.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) failSink(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Count int `json:"count"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.engine.Sinks.SetFailures(chi.URLParam(r, "id"), req.Count); err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "armed"})
}

func (s *Server) listPipelines(w http.ResponseWriter, r *http.Request) {
	all := s.engine.Pipelines.All()
	writeJSON(w, http.StatusOK, all)
}

func (s *Server) switchPipeline(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SourceID  string   `json:"source_id"`
		BatchSize int      `json:"batch_size"`
		SinkIDs   []string `json:"sink_ids"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	cfg, err := s.engine.SwitchPipeline(req.SourceID, req.BatchSize, req.SinkIDs)
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, cfg)
}

func (s *Server) getRules(w http.ResponseWriter, r *http.Request) {
	ruleset, err := s.engine.Rules.Current()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, ruleset)
}

func (s *Server) updateRules(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Rules []config.RoutingRule `json:"rules"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	ruleset, err := s.engine.UpdateRules(req.Rules)
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, ruleset)
}

func (s *Server) listDeadLetters(w http.ResponseWriter, r *http.Request) {
	all, err := s.engine.DeadLetters.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, all)
}

func (s *Server) requeueDeadLetter(w http.ResponseWriter, r *http.Request) {
	letter, rule, err := s.engine.RequeueDead(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deadletter": letter, "rule": rule})
}

func (s *Server) auditList(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	entries, err := s.engine.Audit.List(limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

func nowNanos() int64 {
	return timeNow().UnixNano()
}
