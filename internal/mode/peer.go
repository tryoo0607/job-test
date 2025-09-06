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
	if cfg.JobIndex == nil {
		return fmt.Errorf("job-index missing in indexed-peer mode")
	}

	idx := *cfg.JobIndex
	targetIndex := (idx + 1) % cfg.TotalPods // 다음 노드

	// 호스트명: myjob-1.subdomain.svc.cluster.local
	targetHost := fmt.Sprintf("myjob-%d.%s", targetIndex, cfg.Subdomain)
	url := fmt.Sprintf("http://%s:8080/ping", targetHost)

	fmt.Printf("[Indexed-Peer] Pod-%d → POST to %s\n", idx, url)

	// 샘플 바디
	body := []byte(fmt.Sprintf(`{"from": %d}`, idx))

	// 간단한 HTTP 요청
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

	fmt.Printf("[Indexed-Peer] Got response: %s\n", resp.Status)

	return nil
}
