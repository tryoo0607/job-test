package mode

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/tryoo0607/job-test/internal/config"
)

func RunIndexedPeer(ctx context.Context, cfg config.Config) error {
	fmt.Println("🚀 [Peer] Indexed Job 호스트네임 기반 통신 시작")

	// 필수값 체크
	if cfg.JobIndex == nil {
		return fmt.Errorf("JOB_COMPLETION_INDEX missing")
	}
	if cfg.TotalPods <= 0 {
		return fmt.Errorf("total-pods must be > 0")
	}
	if cfg.Subdomain == "" {
		return fmt.Errorf("subdomain (headless svc name) missing")
	}
	if cfg.JobName == "" {
		return fmt.Errorf("job-name missing")
	}

	selfIdx := *cfg.JobIndex
	targetIdx := (selfIdx + 1) % cfg.TotalPods

	// <job-name>-<index>.<subdomain>
	targetHost := fmt.Sprintf("%s-%d.%s", cfg.JobName, targetIdx, cfg.Subdomain)
	url := fmt.Sprintf("http://%s:8080/ping", targetHost)
	fmt.Printf("🎯 target: %s (self=%d → next=%d)\n", url, selfIdx, targetIdx)

	// (권장) 짧은 재시도: DNS/Endpoint 전파 지연 흡수
	client := &http.Client{Timeout: 3 * time.Second}
	var lastErr error
	for attempt := 1; attempt <= 5; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBufferString(
			fmt.Sprintf(`{"from":%d}`, selfIdx),
		))
		if err != nil {
			return fmt.Errorf("build request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err == nil && resp != nil {
			defer resp.Body.Close()
			fmt.Printf("✅ 응답: %s (attempt=%d)\n", resp.Status, attempt)
			return nil
		}
		lastErr = err
		time.Sleep(time.Duration(attempt) * 200 * time.Millisecond)
	}
	return fmt.Errorf("send to %s failed: %v", targetHost, lastErr)
}
