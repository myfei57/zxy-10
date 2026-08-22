package engine

import (
	"path/filepath"

	"edgelog/internal/config"
)

// SeedDemo registers a small demonstration pipeline so the console pages have
// content on first start. It exercises every domain component through its
// real call chains.
func (e *Engine) SeedDemo() error {
	existing, err := e.Sinks.List()
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil
	}
	primary, err := e.Sinks.Register("主日志中心")
	if err != nil {
		return err
	}
	archive, err := e.Sinks.Register("归档接收端")
	if err != nil {
		return err
	}
	source, err := e.RegisterSource("网关-01", filepath.Join(e.cfg.DataDir, "demo-access.log"))
	if err != nil {
		return err
	}
	if _, err := e.Rules.Update([]config.RoutingRule{
		{Topic: source.ID, SinkID: primary.ID, Partition: 1},
		{Topic: source.ID + "-audit", SinkID: archive.ID, Partition: 2},
	}); err != nil {
		return err
	}
	if _, err := e.SwitchPipeline(source.ID, e.cfg.BatchSize, []string{primary.ID}); err != nil {
		return err
	}
	if _, err := e.Subs.Bind(source.ID, primary.ID); err != nil {
		return err
	}
	if err := e.Ingest(source.ID, []string{
		"GET /api/health 200",
		"POST /api/login 401",
		"GET /api/users 200",
	}); err != nil {
		return err
	}
	if err := e.Tick(e.Now()); err != nil {
		return err
	}
	_, _ = e.Audit.Record("system", "demo.seed", "network", "demo", "演示管道初始化完成")
	return nil
}
