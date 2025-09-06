package mode

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"sort"
	"time"

	"github.com/tryoo0607/job-test/internal/config"
	"github.com/tryoo0607/job-test/internal/runtime"
)

func RunIndexedPeer(ctx context.Context, cfg config.Config) error {
	fmt.Println("🚀 [Peer] Headless SVC DNS 기반 링 통신 시작")
	if cfg.JobIndex == nil {
		return fmt.Errorf("JOB_COMPLETION_INDEX missing")
	}
	if cfg.Subdomain == "" {
		return fmt.Errorf("subdomain (headless svc name) missing")
	}

	myIdx := *cfg.JobIndex

	// 수신 서버는 먼저 띄워둠
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("pong"))
		})
		_ = http.ListenAndServe(":8080", mux)
	}()

	svc := cfg.Subdomain

	// 1) WARM-UP: peers 모일 때까지 대기
	// want 계산: TotalPods를 넘겨주고 있으면 그 값, 아니면 최소 2
	want := 2
	if cfg.TotalPods > 0 {
		want = cfg.TotalPods
	}

	warmup := 30 * time.Second // 기본값 (원하면 config로 뺄 수 있음)
	peers, err := waitForPeers(ctx, svc, want, warmup)
	if err != nil {
		return err
	}
	fmt.Printf("🔎 peers(%d): %v\n", len(peers), peers)

	targetIP := peers[(myIdx+1)%len(peers)]
	url := fmt.Sprintf("http://%s:8080/ping", targetIP)
	fmt.Printf("🎯 target: %s (selfIdx=%d)\n", url, myIdx)

	// 2) HTTP 재시도: runtime.WithRetry 사용 (cfg.RetryMax / cfg.RetryBackoff로 조절)
	client := &http.Client{Timeout: 5 * time.Second}
	err = runtime.WithRetry(ctx, cfg.RetryMax, cfg.RetryBackoff, func() error {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBufferString(
			fmt.Sprintf(`{"from":%d}`, myIdx),
		))
		if err != nil {
			return fmt.Errorf("build request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			fmt.Printf("✅ 응답: %s\n", resp.Status)
			// 성공 시에만 hold
			runtime.Hold(ctx, cfg.HoldFor)
			return nil
		}
		return fmt.Errorf("bad status: %s", resp.Status)
	})
	if err != nil {
		return fmt.Errorf("send to %s failed: %v", url, err)
	}
	return nil
}

func waitForPeers(ctx context.Context, svc string, want int, warmup time.Duration) ([]string, error) {
	// want: 최소 필요한 peers 수 (예: 2 또는 cfg.TotalPods)
	deadline, cancel := context.WithTimeout(ctx, warmup)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline.Done():
			return nil, fmt.Errorf("peer warmup timeout after %s (want=%d)", warmup, want)
		case <-ticker.C:
			// DNS 조회
			addrs, _ := net.DefaultResolver.LookupIPAddr(deadline, svc)
			if len(addrs) >= want {
				peers := make([]string, 0, len(addrs))
				for _, a := range addrs {
					peers = append(peers, a.IP.String())
				}
				sort.Strings(peers)
				return peers, nil
			}
		}
	}
}
