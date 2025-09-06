package mode

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/tryoo0607/job-test/internal/config"
)

func RunIndexedPeer(ctx context.Context, cfg config.Config) error {
	fmt.Println("🚀 [Peer] Deployment 모드: DNS로 피어 탐색 시작")

	selfIP := os.Getenv("POD_IP") // Downward API로 주입
	if selfIP == "" {
		return fmt.Errorf("POD_IP env missing (set via fieldRef:status.podIP)")
	}

	// 헤드리스 서비스 이름(= cfg.Subdomain)로 A레코드 조회
	svc := cfg.Subdomain // 예: "peer-svc"
	ips, err := net.LookupIP(svc)
	if err != nil {
		return fmt.Errorf("lookup %s failed: %w", svc, err)
	}
	if len(ips) == 0 {
		return fmt.Errorf("no peers found for %s", svc)
	}

	// 문자열로 정렬(결정적 순서)
	var peers []string
	for _, ip := range ips {
		peers = append(peers, ip.String())
	}
	sort.Strings(peers)
	fmt.Printf("🔎 peers: %v\n", peers)

	// 내 위치 찾기
	selfIdx := -1
	for i, ip := range peers {
		if ip == selfIP {
			selfIdx = i
			break
		}
	}
	if selfIdx == -1 {
		return fmt.Errorf("self IP %s not in peers %v", selfIP, peers)
	}
	if len(peers) == 1 {
		fmt.Println("ℹ️ 단일 파드 환경 → 스킵")
		return nil
	}

	targetIdx := (selfIdx + 1) % len(peers)
	targetIP := peers[targetIdx]
	url := fmt.Sprintf("http://%s:8080/ping", targetIP)

	fmt.Printf("🌐 POST → %s (from %s)\n", url, selfIP)

	body := []byte(fmt.Sprintf(`{"from":"%s"}`, selfIP))
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
