package mode

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
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

	rdb := redis.NewClient(&redis.Options{Addr: mustParseRedis(cfg.QueueURL)})
	defer rdb.Close()

	// 0) 시작 시 큐가 비었으면 즉시 Complete (hold 없음)
	llen, err := rdb.LLen(ctx, cfg.QueueKey).Result()
	if err != nil {
		return fmt.Errorf("redis LLEN error: %w", err)
	}
	if llen == 0 {
		fmt.Printf("🟢 [Queue] '%s'가 시작 시 비어 있음 → 즉시 Completed\n", cfg.QueueKey)
		return nil
	}

	p := processor.UppercaseFile{OutDir: cfg.OutputDir}
	fmt.Printf("📤 [Queue] Processor 출력 디렉토리: %s\n", cfg.OutputDir)

	// 성공 개수 카운트 (성공시에만 hold)
	var processedOK atomic.Int64

	g, gctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(gctx)
	defer cancel()

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

				// BRPOP: timeout 나면 큐가 비었다고 보고 전체 종료
				res, err := rdb.BRPop(ctx, 3*time.Second, cfg.QueueKey).Result()
				if err == redis.Nil {
					fmt.Printf("🟡 [Worker-%d] BRPOP timeout → 큐 비었다고 판단, 전체 종료\n", workerID)
					cancel()
					return nil
				}
				if err != nil {
					fmt.Printf("❌ [Worker-%d] Redis pop 에러: %v\n", workerID, err)
					return fmt.Errorf("redis pop error: %w", err)
				}

				payload := res[1]
				it := processor.Item{ID: generateID(payload), Path: payload}
				fmt.Printf("📥 [Worker-%d] 처리할 항목 수신: ID=%s, Path=%s\n", workerID, it.ID, it.Path)

				err = runtime.WithRetry(ctx, cfg.RetryMax, cfg.RetryBackoff, func() error {
					fmt.Printf("⚙️  [Worker-%d] Processor 실행 시작 (ID=%s)\n", workerID, it.ID)
					return p.Process(ctx, it)
				})
				if err != nil {
					fmt.Printf("❌ [Worker-%d] 처리 실패 (ID=%s): %v\n", workerID, it.ID, err)
					continue
				}

				processedOK.Add(1)
				fmt.Printf("✅ [Worker-%d] 처리 성공 (누계=%d)\n", workerID, processedOK.Load())
			}
		})
	}

	// 워커 종료 대기
	err = g.Wait()

	// ✅ 성공시에만 hold (성공 처리 건수가 1개 이상이고, 전체 에러가 없는 경우)
	if err == nil && processedOK.Load() > 0 && cfg.HoldFor > 0 {
		fmt.Printf("⏳ [Queue] 성공 처리 감지 → hold for %s...\n", cfg.HoldFor)
		runtime.Hold(ctx, cfg.HoldFor)
	}

	return err
}

// Redis URL 파싱
func mustParseRedis(url string) string {
	if strings.HasPrefix(url, "redis://") {
		return strings.TrimPrefix(url, "redis://")
	}
	return url
}

// ID 생성
func generateID(path string) string {
	parts := strings.Split(path, "/")
	return strings.TrimSuffix(parts[len(parts)-1], ".txt")
}
