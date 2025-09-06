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
	fmt.Println("🚀 [Peer] Deployment 모드: DNS로 피어 탐색 시작")

	// ✅ config에서 주입
	if cfg.PodIP == "" {
		return fmt.Errorf("POD_IP (cfg.PodIP) missing; set via Downward API (status.podIP)")
	}
	if cfg.Subdomain == "" {
		return fmt.Errorf("SUBDOMAIN (cfg.Subdomain) missing; set headless svc name")
	}

	// ✅ 컨텍스트 기반 DNS 조회(타임아웃)
	dnsCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	addrs, err := net.DefaultResolver.LookupIPAddr(dnsCtx, cfg.Subdomain)
	if err != nil {
		return fmt.Errorf("lookup %s failed: %w", cfg.Subdomain, err)
	}
	if len(addrs) == 0 {
		return fmt.Errorf("no peers found for %s", cfg.Subdomain)
	}

	// 문자열로 정렬(결정적 순서)
	peers := make([]string, 0, len(addrs))
	for _, a := range addrs {
		peers = append(peers, a.IP.String())
	}
	sort.Strings(peers)
	fmt.Printf("🔎 peers: %v\n", peers)

	// 내 위치 찾기
	selfIdx := -1
	for i, ip := range peers {
		if ip == cfg.PodIP {
			selfIdx = i
			break
		}
	}
	if selfIdx == -1 {
		return fmt.Errorf("self IP %s not in peers %v", cfg.PodIP, peers)
	}
	if len(peers) == 1 {
		fmt.Println("ℹ️ 단일 파드 환경 → 스킵")
		return nil
	}

	targetIdx := (selfIdx + 1) % len(peers)
	targetIP := peers[targetIdx]
	url := fmt.Sprintf("http://%s:8080/ping", targetIP)

	fmt.Printf("🌐 POST → %s (from %s)\n", url, cfg.PodIP)

	body := []byte(fmt.Sprintf(`{"from":"%s"}`, cfg.PodIP))
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("✅ 응답: %s\n", resp.Status)
	return nil
}
