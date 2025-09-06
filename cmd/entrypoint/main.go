package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/tryoo0607/job-test/internal/api"
	"github.com/tryoo0607/job-test/internal/config"
	"github.com/tryoo0607/job-test/internal/mode"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var errRun error
	switch cfg.Mode {
	case api.Indexed:
		errRun = mode.RunIndexed(ctx, *cfg)
	case api.Peer:
		errRun = mode.RunIndexedPeer(ctx, *cfg)
	case api.Fixed:
		errRun = mode.RunFixed(ctx, *cfg)
	case api.Queue:
		errRun = mode.RunQueue(ctx, *cfg)
	default:
		errRun = fmt.Errorf("unknown mode: %s", cfg.Mode)
	}
	if errRun != nil {
		log.Fatal(errRun)
	}
}
