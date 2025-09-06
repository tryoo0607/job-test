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
	fmt.Println("🚀 [main] 프로그램 시작")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ [main] 설정 로딩 실패: %v\n", err)
	}

	fmt.Printf("🔧 [main] 설정 로딩 완료: Mode=%s\n", cfg.Mode)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var errRun error
	switch cfg.Mode {
	case api.Indexed:
		fmt.Println("⚙️  [main] Indexed 모드 진입")
		errRun = mode.RunIndexed(ctx, *cfg)
	case api.Peer:
		fmt.Println("⚙️  [main] IndexedPeer 모드 진입")
		errRun = mode.RunIndexedPeer(ctx, *cfg)
	case api.Fixed:
		fmt.Println("⚙️  [main] Fixed 모드 진입")
		errRun = mode.RunFixed(ctx, *cfg)
	case api.Queue:
		fmt.Println("⚙️  [main] Queue 모드 진입")
		errRun = mode.RunQueue(ctx, *cfg)
	default:
		errRun = fmt.Errorf("unknown mode: %s", cfg.Mode)
	}

	if errRun != nil {
		log.Fatalf("❌ [main] 실행 실패: %v\n", errRun)
	} else {
		fmt.Println("✅ [main] 작업 정상 종료")
	}
}
