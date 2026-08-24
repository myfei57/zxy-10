package config

import (
	"errors"
	"time"

	"edgelog/internal/audit"
	"edgelog/internal/store"
)

// PipelineConfig describes one log collection pipeline: which source it
// tails, the batch size and the sinks it feeds.
type PipelineConfig struct {
	SourceID  string   `json:"source_id"`
	BatchSize int      `json:"batch_size"`
	SinkIDs   []string `json:"sink_ids"`
	Active    bool     `json:"active"`
	Updated   string   `json:"updated"`
}

// PipelineSwapper is implemented by the tailer pipeline manager.
type PipelineSwapper interface {
	Attach(sourceID string, cfg PipelineConfig) error
	Detach(sourceID string) error
	Active(sourceID string) (PipelineConfig, bool)
}

// SwitchService hot-swaps collection pipelines. The new configuration is
// durably stored before the old pipeline is torn down, so a failed config
// write never leaves a source without a consumer.
type SwitchService struct {
	fs      *store.FileStore
	swapper PipelineSwapper
	audit   audit.Recorder
	clock   func() time.Time
}

// NewSwitchService creates the pipeline switch service.
func NewSwitchService(fs *store.FileStore, swapper PipelineSwapper) *SwitchService {
	return &SwitchService{fs: fs, swapper: swapper, clock: time.Now}
}

// NewSwitchServiceWithAudit wires the audit sink into the switch service.
func NewSwitchServiceWithAudit(fs *store.FileStore, swapper PipelineSwapper, recorder audit.Recorder) *SwitchService {
	s := NewSwitchService(fs, swapper)
	s.audit = recorder
	return s
}

func (s *SwitchService) now() string {
	return s.clock().UTC().Format(time.RFC3339)
}

// Path resolves the durable pipeline config of a source.
func (s *SwitchService) Path(sourceID string) string {
	return s.fs.Path("pipelines", sourceID)
}

// Switch installs a new pipeline configuration. The config is written
// durably first; only then is the old pipeline detached and the new one
// attached.
func (s *SwitchService) Switch(sourceID string, cfg PipelineConfig) (PipelineConfig, error) {
	if sourceID == "" {
		return PipelineConfig{}, errors.New("source id is required")
	}
	cfg.Active = true
	cfg.Updated = s.now()
	if s.swapper != nil {
		if err := s.swapper.Detach(sourceID); err != nil {
			return PipelineConfig{}, err
		}
		if err := s.swapper.Attach(sourceID, cfg); err != nil {
			return PipelineConfig{}, err
		}
	}
	if err := s.fs.WriteJSON(s.Path(sourceID), cfg); err != nil {
		return PipelineConfig{}, err
	}
	if s.audit != nil {
		_, _ = s.audit.Record("console", "pipeline.switch", "source", sourceID, "batch="+itoa(cfg.BatchSize))
	}
	return cfg, nil
}
