package runtime

import (
	"context"

	"github.com/tryoo0607/job-test/internal/processor"
	"golang.org/x/sync/errgroup"
)

func RunPool(ctx context.Context, maxConcurrency int, ch <-chan processor.Item, p processor.Processor) error {

	g, ctx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, maxConcurrency)

	for it := range ch {
		it := it // capture loop variable
		sem <- struct{}{}
		g.Go(func() error {
			defer func() { <-sem }()
			return p.Process(ctx, it)
		})
	}

	return g.Wait()
}
