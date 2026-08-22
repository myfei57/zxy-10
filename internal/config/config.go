package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config carries the runtime settings for the EdgeLog control plane. Every
// value can be overridden through environment variables so the same binary
// works for local development and the containerized delivery image.
type Config struct {
	ListenAddr         string
	DataDir            string
	TickSeconds        int
	BatchSize          int
	RetryBudget        int
	DedupWindowSeconds int
	QuotaWindowSeconds int
}

// Default returns the baseline configuration used when no environment
// override is present.
func Default() Config {
	return Config{
		ListenAddr:         ":8080",
		DataDir:            "data",
		TickSeconds:        1,
		BatchSize:          100,
		RetryBudget:        3,
		DedupWindowSeconds: 60,
		QuotaWindowSeconds: 60,
	}
}

// Load reads the environment and applies any overrides on top of Default.
func Load() (Config, error) {
	cfg := Default()
	var err error
	if cfg.ListenAddr, err = envString("EDGELOG_LISTEN", cfg.ListenAddr); err != nil {
		return Config{}, err
	}
	if cfg.DataDir, err = envString("EDGELOG_DATA", cfg.DataDir); err != nil {
		return Config{}, err
	}
	if cfg.TickSeconds, err = envInt("EDGELOG_TICK_SECONDS", cfg.TickSeconds); err != nil {
		return Config{}, err
	}
	if cfg.BatchSize, err = envInt("EDGELOG_BATCH_SIZE", cfg.BatchSize); err != nil {
		return Config{}, err
	}
	if cfg.RetryBudget, err = envInt("EDGELOG_RETRY_BUDGET", cfg.RetryBudget); err != nil {
		return Config{}, err
	}
	if cfg.DedupWindowSeconds, err = envInt("EDGELOG_DEDUP_WINDOW", cfg.DedupWindowSeconds); err != nil {
		return Config{}, err
	}
	if cfg.QuotaWindowSeconds, err = envInt("EDGELOG_QUOTA_WINDOW", cfg.QuotaWindowSeconds); err != nil {
		return Config{}, err
	}
	if cfg.TickSeconds <= 0 || cfg.BatchSize <= 0 {
		return Config{}, fmt.Errorf("tick seconds and batch size must be positive")
	}
	return cfg, nil
}

func envString(name, fallback string) (string, error) {
	if value := os.Getenv(name); value != "" {
		return value, nil
	}
	return fallback, nil
}

func envInt(name string, fallback int) (int, error) {
	raw := os.Getenv(name)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", name, err)
	}
	return value, nil
}
