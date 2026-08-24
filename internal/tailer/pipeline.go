package tailer

import (
	"edgelog/internal/config"
)

// PipelineManager tracks the active collection pipelines per source.
type PipelineManager struct {
	active map[string]config.PipelineConfig
}

// NewPipelineManager creates an empty pipeline manager.
func NewPipelineManager() *PipelineManager {
	return &PipelineManager{active: make(map[string]config.PipelineConfig)}
}

// Attach registers a source pipeline as active.
func (p *PipelineManager) Attach(sourceID string, cfg config.PipelineConfig) error {
	p.active[sourceID] = cfg
	return nil
}

// Detach removes a source pipeline.
func (p *PipelineManager) Detach(sourceID string) error {
	delete(p.active, sourceID)
	return nil
}

// Active returns the active pipeline of a source, if any.
func (p *PipelineManager) Active(sourceID string) (config.PipelineConfig, bool) {
	cfg, ok := p.active[sourceID]
	return cfg, ok
}

// All returns every active pipeline keyed by source.
func (p *PipelineManager) All() map[string]config.PipelineConfig {
	out := make(map[string]config.PipelineConfig, len(p.active))
	for sourceID, cfg := range p.active {
		out[sourceID] = cfg
	}
	return out
}
