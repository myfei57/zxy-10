package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"edgelog/internal/config"
	"edgelog/internal/console"
	"edgelog/internal/engine"
)

func main() {
	addr := flag.String("addr", "", "listen address (overrides EDGELOG_LISTEN)")
	dataDir := flag.String("data", "", "data directory (overrides EDGELOG_DATA)")
	demo := flag.Bool("demo", false, "seed demonstration data when the store is empty")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if *addr != "" {
		cfg.ListenAddr = *addr
	}
	if *dataDir != "" {
		cfg.DataDir = *dataDir
	}

	eng, err := engine.Build(cfg)
	if err != nil {
		log.Fatalf("build engine: %v", err)
	}
	if *demo {
		if err := eng.SeedDemo(); err != nil {
			log.Fatalf("seed demo data: %v", err)
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		if err := eng.Run(ctx); err != nil {
			log.Printf("engine stopped: %v", err)
		}
	}()

	server := console.NewServer(eng)
	log.Printf("EdgeLog listening on %s, data dir %s", cfg.ListenAddr, cfg.DataDir)
	if err := server.ListenAndServe(cfg.ListenAddr); err != nil && err != http.ErrServerClosed {
		log.Fatalf("http server: %v", err)
	}
}
