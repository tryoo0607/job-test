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
	fmt.Println("🚀 [Queue] RunQueue 시작")
	fmt.Printf("🔧 [Queue] Redis URL: %s, Queue Key: %s\n", cfg.QueueURL, cfg.QueueKey)
	fmt.Printf("🔧 [Queue] MaxConcurrency: %d, RetryMax: %d, RetryBackoff: %s\n", cfg.MaxConcurrency, cfg.RetryMax, cfg.RetryBackoff)

	rdb := redis.NewClient(&redis.Options{
		Addr: mustParseRedis(cfg.QueueURL),
	})
	defer rdb.Close()

	p := processor.UppercaseFile{OutDir: cfg.OutputDir}
	fmt.Printf("📤 [Queue] Processor 출력 디렉토리: %s\n", cfg.OutputDir)

	g, ctx := errgroup.WithContext(ctx)

	for i := 0; i < cfg.MaxConcurrency; i++ {
		workerID := i
		fmt.Printf("👷 [Queue] 워커 #%d 시작\n", workerID)

		g.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					fmt.Printf("🛑 [Worker-%d] 컨텍스트 종료됨\n", workerID)
					return ctx.Err()
				default:
				}

				// 큐에서 항목 하나 꺼내기 (blocking pop)
				res, err := rdb.BRPop(ctx, 5*time.Second, cfg.QueueKey).Result()
				if err == redis.Nil {
					fmt.Printf("⌛ [Worker-%d] 대기 중... (큐 비어있음)\n", workerID)
					continue
				}
				if err != nil {
					fmt.Printf("❌ [Worker-%d] Redis pop 에러: %v\n", workerID, err)
					return fmt.Errorf("redis pop error: %w", err)
				}

				payload := res[1] // res[0]은 queue key
				it := processor.Item{
					ID:   generateID(payload),
					Path: payload,
				}

				fmt.Printf("📥 [Worker-%d] 처리할 항목 수신: ID=%s, Path=%s\n", workerID, it.ID, it.Path)

				err = runtime.WithRetry(ctx, cfg.RetryMax, cfg.RetryBackoff, func() error {
					fmt.Printf("⚙️  [Worker-%d] Processor 실행 시작 (ID=%s)\n", workerID, it.ID)
					return p.Process(ctx, it)
				})

				if err != nil {
					fmt.Printf("❌ [Worker-%d] 처리 실패 (ID=%s): %v\n", workerID, it.ID, err)
					// TODO: DLQ 전송 또는 재큐잉 등 정책
				} else {
					fmt.Printf("✅ [Worker-%d] 처리 성공 (ID=%s)\n", workerID, it.ID)
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
