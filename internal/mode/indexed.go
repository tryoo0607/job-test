package mode

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/tryoo0607/job-test/internal/config"
	"github.com/tryoo0607/job-test/internal/processor"
	"github.com/tryoo0607/job-test/internal/runtime"
)

func RunIndexed(ctx context.Context, cfg config.Config) error {

	if cfg.JobIndex == nil {
		return fmt.Errorf("job-index missing in indexed mode")
	}

	idx := *cfg.JobIndex

	in := filepath.Join(cfg.InputDir, fmt.Sprintf("input-%d.txt", idx))

	p := processor.UppercaseFile{OutDir: cfg.OutputDir}

	return runtime.WithRetry(ctx, cfg.RetryMax, cfg.RetryBackoff, func() error {
		return p.Process(ctx, processor.Item{ID: strconv.Itoa(idx), Path: in})
	})
}
