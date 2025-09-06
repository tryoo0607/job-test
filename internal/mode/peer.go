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

	// (옵션) 수신 서버: 다른 파드가 날 /ping 할 수 있게 열어둠
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("pong"))
		})
		_ = http.ListenAndServe(":8080", mux)
	}()

	svc := cfg.Subdomain // 같은 네임스페이스면 svc 짧은 이름으로도 OK

	// DNS 조회 재시도(엔드포인트 전파 지연 흡수)
	var addrs []net.IPAddr
	for attempt := 1; attempt <= 6; attempt++ { // 최대 ~9초 대기
		dnsCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		a, err := net.DefaultResolver.LookupIPAddr(dnsCtx, svc)
		cancel()
		if err == nil && len(a) > 0 {
			addrs = a
			break
		}
		time.Sleep(time.Duration(attempt) * 300 * time.Millisecond)
	}
	if len(addrs) == 0 {
		return fmt.Errorf("dns lookup failed for %s (no endpoints)", svc)
	}

	// peers 정렬(모든 파드에서 결정적 순서 보장)
	peers := make([]string, 0, len(addrs))
	for _, a := range addrs {
		peers = append(peers, a.IP.String())
	}
	sort.Strings(peers)
	fmt.Printf("🔎 peers(%d): %v\n", len(peers), peers)

	// 내 다음 타겟 = (myIdx+1) % len(peers)
	targetIP := peers[(myIdx+1)%len(peers)]
	url := fmt.Sprintf("http://%s:8080/ping", targetIP)
	fmt.Printf("🎯 target: %s (selfIdx=%d)\n", url, myIdx)

	client := &http.Client{Timeout: 3 * time.Second}
	var lastErr error
	for attempt := 1; attempt <= 5; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBufferString(
			fmt.Sprintf(`{"from":%d}`, myIdx),
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
	return fmt.Errorf("send to %s failed: %v", url, lastErr)
}
