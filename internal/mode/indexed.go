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
	fmt.Println("🚀 [Indexed] RunIndexed 시작")

	if cfg.JobIndex == nil {
		return fmt.Errorf("job-index missing in indexed mode")
	}

	idx := *cfg.JobIndex
	fmt.Printf("🔧 [Indexed] 받은 job-index: %d\n", idx)

	in := filepath.Join(cfg.InputDir, fmt.Sprintf("input-%02d.txt", idx))
	fmt.Printf("📄 [Indexed] 처리할 입력 파일 경로: %s\n", in)

	p := processor.UppercaseFile{OutDir: cfg.OutputDir}
	fmt.Printf("📤 [Indexed] 출력 디렉토리: %s\n", cfg.OutputDir)

	err := runtime.WithRetry(ctx, cfg.RetryMax, cfg.RetryBackoff, func() error {
		fmt.Println("⚙️  [Indexed] Processor 실행 시작")
		return p.Process(ctx, processor.Item{ID: strconv.Itoa(idx), Path: in})
	})

	if err != nil {
		fmt.Printf("❌ [Indexed] Processor 실행 실패: %v\n", err)
	} else {
		fmt.Println("✅ [Indexed] Processor 실행 완료")
		runtime.Hold(ctx, cfg.HoldFor)
	}

	return err
}
