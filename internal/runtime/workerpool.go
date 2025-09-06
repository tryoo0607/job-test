package runtime

import (
	"context"
	"fmt"

	"github.com/tryoo0607/job-test/internal/processor"
	"golang.org/x/sync/errgroup"
)

func RunPool(ctx context.Context, maxConcurrency int, ch <-chan processor.Item, p processor.Processor) error {
	fmt.Printf("🔧 [RunPool] 고루틴 풀 시작. maxConcurrency=%d\n", maxConcurrency)

	g, ctx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, maxConcurrency)

	for it := range ch {
		it := it // capture loop variable
		fmt.Printf("⚙️  [RunPool] 처리 시작: ID=%s, Path=%s\n", it.ID, it.Path)

		sem <- struct{}{}
		g.Go(func() error {
			defer func() { <-sem }()
			err := p.Process(ctx, it)
			if err != nil {
				fmt.Printf("❌ [RunPool] 처리 실패: ID=%s, 에러=%v\n", it.ID, err)
			} else {
				fmt.Printf("✅ [RunPool] 처리 성공: ID=%s\n", it.ID)
			}
			return err
		})
	}

	err := g.Wait()
	if err != nil {
		fmt.Printf("❌ [RunPool] 전체 작업 중 오류 발생: %v\n", err)
	} else {
		fmt.Println("✅ [RunPool] 전체 작업 완료")
	}
	return err
}
