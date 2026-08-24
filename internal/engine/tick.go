package engine

import (
	"context"
	"os"
	"time"

	"edgelog/internal/config"
	"edgelog/internal/store"
	"edgelog/internal/tailer"
)

func openAppend(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
}

// Run starts the pipeline tick loop and blocks until the context is cancelled.
func (e *Engine) Run(ctx context.Context) error {
	sources, err := e.Tailer.List()
	if err != nil {
		return err
	}
	for _, source := range sources {
		if _, err := e.Batches.Restore(source.ID); err != nil {
			return err
		}
	}
	ticker := time.NewTicker(time.Duration(e.cfg.TickSeconds) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case now := <-ticker.C:
			if err := e.Tick(now); err != nil {
				return err
			}
		}
	}
}

// Tick advances every active pipeline one step.
func (e *Engine) Tick(now time.Time) error {
	for sourceID, pipeline := range e.Pipelines.All() {
		if err := e.stepSource(sourceID, pipeline, now); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) stepSource(sourceID string, pipeline config.PipelineConfig, now time.Time) error {
	source, err := e.Tailer.Get(sourceID)
	if err != nil {
		return err
	}
	current, err := e.Tailer.Current(sourceID)
	if err != nil {
		return err
	}
	rotated, err := tailer.DetectRotation(source.Path, current.Offset)
	if err == nil && rotated {
		current, err = e.Tailer.Reset(sourceID)
		if err != nil {
			return err
		}
	}
	result, err := tailer.ReadLines(source.Path, current.Offset)
	if err != nil {
		return err
	}
	if len(result.Lines) == 0 {
		return nil
	}
	allowed := e.Quota.Allow(sourceID)
	if !allowed {
		e.Flow.SetThrottled(sourceID, true)
		return nil
	}
	if e.Flow.IsThrottled(sourceID) {
		e.Quota.Recover(sourceID)
	}
	if e.Flow.ShouldThrottle(sourceID, e.Router.Buffer(sourceID)) {
		return nil
	}
	accepted := make([]string, 0, len(result.Lines))
	for _, line := range result.Lines {
		if e.dedup.Accept(sourceID, line) {
			accepted = append(accepted, line)
		}
	}
	buf := e.Router.Buffer(sourceID)
	buf.Write(accepted)
	lines := buf.Read(pipeline.BatchSize)
	if len(lines) == 0 {
		return nil
	}
	record, err := e.Batches.Assemble(sourceID, sourceID, lines, result.Offset)
	if err != nil {
		return err
	}
	committed, err := e.Batches.Commit(sourceID, record, result.Offset, current.Line+len(result.Lines))
	if err != nil {
		return err
	}
	rule, err := e.Router.PartitionForTopic(committed.Topic)
	if err != nil {
		return err
	}
	delivered, err := e.sinkRetry(committed, rule.SinkID)
	if err != nil {
		return err
	}
	if delivered.State == store.BatchDead {
		_, _ = e.DeadLetters.Hold(delivered.ID, sourceID, committed.Topic, rule.SinkID, rule.Partition, "retry budget exhausted")
	}
	return nil
}
