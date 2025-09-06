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
	fmt.Println("🚀 [Peer] RunIndexedPeer 시작")

	if cfg.JobIndex == nil {
		return fmt.Errorf("job-index missing in indexed-peer mode")
	}

	idx := *cfg.JobIndex
	fmt.Printf("🔧 [Peer] 현재 job-index: %d\n", idx)
	fmt.Printf("🔧 [Peer] 전체 Pod 수: %d\n", cfg.TotalPods)

	targetIndex := (idx + 1) % cfg.TotalPods
	fmt.Printf("🎯 [Peer] 타겟 인덱스: %d\n", targetIndex)

	targetHost := fmt.Sprintf("myjob-%d.%s", targetIndex, cfg.Subdomain)
	url := fmt.Sprintf("http://%s:8080/ping", targetHost)

	fmt.Printf("🌐 [Peer] HTTP POST 전송: %s\n", url)

	body := []byte(fmt.Sprintf(`{"from": %d}`, idx))

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send request to %s: %w", targetHost, err)
	}
	defer resp.Body.Close()

	fmt.Printf("✅ [Peer] 응답 수신 완료: %s\n", resp.Status)

	return nil
}
