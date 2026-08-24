package engine

import (
	"fmt"
	"time"

	"edgelog/internal/audit"
	"edgelog/internal/batch"
	"edgelog/internal/buffer"
	"edgelog/internal/config"
	"edgelog/internal/deadletter"
	"edgelog/internal/dedup"
	"edgelog/internal/quota"
	"edgelog/internal/router"
	"edgelog/internal/sink"
	"edgelog/internal/store"
	"edgelog/internal/tailer"
)

// Engine wires every pipeline component together and drives the tick loop:
// read, buffer, assemble, commit, dispatch, retry and dead-letter.
type Engine struct {
	cfg         config.Config
	Audit       *audit.Service
	Sinks       *sink.Registry
	Rules       *config.RulesService
	Subs        *router.SubscriptionService
	Pipelines   *tailer.PipelineManager
	Switch      *config.SwitchService
	Tailer      *tailer.Service
	Batches     *batch.Service
	Quota       *quota.Service
	Flow        *router.FlowController
	Router      *router.Router
	DeadLetters *deadletter.Service
	dedup       *dedup.Deduper
}

// Build constructs the full engine from configuration, opening the file store
// and wiring the dependency graph in one place.
func Build(cfg config.Config) (*Engine, error) {
	st, err := store.Open(cfg.DataDir)
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}
	fs := st.FS
	recorder := audit.NewService(fs)
	sinks := sink.NewRegistry(fs)
	rules := config.NewRulesServiceWithAudit(fs, recorder)
	flow := router.NewFlowController()
	routerSvc := router.NewRouter(rules, flow)
	subs := router.NewSubscriptionService(fs, rules)
	rules.SetRebinder(subs)
	pipelines := tailer.NewPipelineManager()
	switchSvc := config.NewSwitchServiceWithAudit(fs, pipelines, recorder)
	checkpoints := store.NewCheckpointStore(fs)
	batchesStore := store.NewBatchStore(fs)
	deduper := dedup.NewDeduper(cfg.DedupWindowSeconds)
	tailerSvc := tailer.NewService(fs, checkpoints)
	batchSvc := batch.NewServiceWithAudit(fs, batchesStore, checkpoints, deduper, tailerSvc, recorder)
	quotaSvc := quota.NewServiceWithLimiter(float64(cfg.BatchSize), cfg.QuotaWindowSeconds, flow)
	deadletters := deadletter.NewServiceWithAudit(fs, recorder)
	return &Engine{
		cfg:         cfg,
		Audit:       recorder,
		Sinks:       sinks,
		Rules:       rules,
		Subs:        subs,
		Pipelines:   pipelines,
		Switch:      switchSvc,
		Tailer:      tailerSvc,
		Batches:     batchSvc,
		Quota:       quotaSvc,
		Flow:        flow,
		Router:      routerSvc,
		DeadLetters: deadletters,
		dedup:       deduper,
	}, nil
}

// RegisterSource registers a log source and attaches its ring buffer.
func (e *Engine) RegisterSource(name, path string) (tailer.Source, error) {
	source, err := e.Tailer.Register(name, path)
	if err != nil {
		return tailer.Source{}, err
	}
	e.Router.RegisterBuffer(source.ID, buffer.NewRingBuffer(e.cfg.BatchSize*10))
	return source, nil
}

// SwitchPipeline hot-switches a source pipeline through the durable-first
// switch service.
func (e *Engine) SwitchPipeline(sourceID string, batchSize int, sinkIDs []string) (config.PipelineConfig, error) {
	cfg := config.PipelineConfig{SourceID: sourceID, BatchSize: batchSize, SinkIDs: sinkIDs}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = e.cfg.BatchSize
	}
	return e.Switch.Switch(sourceID, cfg)
}

// UpdateRules replaces routing rules and rebinds subscriptions.
func (e *Engine) UpdateRules(rules []config.RoutingRule) (config.Ruleset, error) {
	return e.Rules.Update(rules)
}

// RequeueDead redelivers a dead letter using the current partition mapping.
func (e *Engine) RequeueDead(id string) (deadletter.DeadLetter, config.RoutingRule, error) {
	return e.DeadLetters.Requeue(id, e.Router.PartitionForTopic)
}

// Ingest appends lines to a source file so the next tick reads them.
func (e *Engine) Ingest(sourceID string, lines []string) error {
	source, err := e.Tailer.Get(sourceID)
	if err != nil {
		return err
	}
	file, err := openAppend(source.Path)
	if err != nil {
		return err
	}
	defer file.Close()
	for _, line := range lines {
		if _, err := file.WriteString(line + "\n"); err != nil {
			return err
		}
	}
	return nil
}

// Now returns the current UTC time for the tick loop.
func (e *Engine) Now() time.Time {
	return time.Now().UTC()
}

// CfgDataDir exposes the configured data directory for source file paths.
func (e *Engine) CfgDataDir() string {
	return e.cfg.DataDir
}
