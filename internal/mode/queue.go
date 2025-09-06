package mode

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tryoo0607/job-test/internal/config"
	"github.com/tryoo0607/job-test/internal/processor"
	"github.com/tryoo0607/job-test/internal/runtime"
	"golang.org/x/sync/errgroup"
)

func RunQueue(ctx context.Context, cfg config.Config) error {
	rdb := redis.NewClient(&redis.Options{
		Addr: mustParseRedis(cfg.QueueURL),
	})
	defer rdb.Close()

	p := processor.UppercaseFile{OutDir: cfg.OutputDir}
	g, ctx := errgroup.WithContext(ctx)

	for i := 0; i < cfg.MaxConcurrency; i++ {
		g.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				// 큐에서 항목 하나 꺼내기 (blocking pop)
				res, err := rdb.BRPop(ctx, 5*time.Second, cfg.QueueKey).Result()
				if err == redis.Nil {
					continue // timeout
				}
				if err != nil {
					return fmt.Errorf("redis pop error: %w", err)
				}

				payload := res[1] // res[0]은 queue key
				it := processor.Item{
					ID:   generateID(payload),
					Path: payload,
				}

				err = runtime.WithRetry(ctx, cfg.RetryMax, cfg.RetryBackoff, func() error {
					return p.Process(ctx, it)
				})

				if err != nil {
					// TODO: DLQ 전송 또는 재큐잉 등 정책에 따라 처리
					fmt.Printf("[Worker] failed to process item %s: %v\n", it.Path, err)
				}
			}
		})
	}

	return g.Wait()
}

// 단순한 Redis URL 파싱기 (예: redis://localhost:6379 → localhost:6379)
func mustParseRedis(url string) string {
	if strings.HasPrefix(url, "redis://") {
		return strings.TrimPrefix(url, "redis://")
	}
	return url
}

// ID 생성 (경로에서 파일명 추출 또는 UUID 사용 가능)
func generateID(path string) string {
	parts := strings.Split(path, "/")
	return strings.TrimSuffix(parts[len(parts)-1], ".txt")
}
