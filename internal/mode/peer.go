package mode

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"regexp"
	"sort"
	"time"

	"github.com/tryoo0607/job-test/internal/config"
	"github.com/tryoo0607/job-test/internal/runtime"
)

func RunIndexedPeer(ctx context.Context, cfg config.Config) error {
	fmt.Println("🚀 [Peer] Headless SVC per-pod FQDN로 링 통신")

	if cfg.JobIndex == nil {
		return fmt.Errorf("JOB_COMPLETION_INDEX missing")
	}
	if cfg.Subdomain == "" {
		return fmt.Errorf("subdomain (headless svc name) missing")
	}

	myIdx := *cfg.JobIndex

	// 수신 서버 먼저 오픈
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("pong"))
		})
		_ = http.ListenAndServe(":8080", mux)
	}()

	// 총 파드 수 결정: 있으면 그대로, 없으면 DNS로 warm-up해서 개수 추정
	total := cfg.TotalPods
	if total <= 0 {
		peers, err := waitForPeers(ctx, cfg.Subdomain, 2, 60*time.Second) // 최소 2개 모일 때까지
		if err != nil {
			return err
		}
		total = len(peers)
		fmt.Printf("🔢 total pods inferred from DNS: %d\n", total)
	}

	next := (myIdx + 1) % total

	// ✅ per-pod FQDN: <job-name>-<index>.<subdomain>
	//   - job-name은 Indexed Job에서 각 Pod의 hostname이 "<job-name>-<index>"로 자동 설정됨
	//   - 내 hostname에서 base만 추출하거나, env로 넘긴 JOB_NAME을 쓰는 둘 중 택1
	hostBase, _ := os.Hostname() // ex) "peer-job-1"
	base := regexp.MustCompile(`^(.*)-\d+$`).ReplaceAllString(hostBase, `$1`)
	targetHost := fmt.Sprintf("%s-%d.%s", base, next, cfg.Subdomain)

	url := fmt.Sprintf("http://%s:8080/ping", targetHost)
	fmt.Printf("🎯 target: %s (selfIdx=%d, total=%d)\n", url, myIdx, total)

	client := &http.Client{Timeout: 5 * time.Second}
	err := runtime.WithRetry(ctx, cfg.RetryMax, cfg.RetryBackoff, func() error {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBufferString(
			fmt.Sprintf(`{"from":%d}`, myIdx),
		))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode/100 == 2 {
			fmt.Printf("✅ 응답: %s\n", resp.Status)
			runtime.Hold(ctx, cfg.HoldFor) // 성공 시에만 hold
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
