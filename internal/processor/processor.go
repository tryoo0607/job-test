package processor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Item struct {
	ID   string
	Path string
}
type Processor interface {
	Process(ctx context.Context, it Item) error
}

type UppercaseFile struct{ OutDir string }

func (u UppercaseFile) Process(ctx context.Context, it Item) error {
	fmt.Printf("🟡 [Processor] 시작: ID=%s, Path=%s\n", it.ID, it.Path)

	// 1) 입력 파일 읽기
	data, err := os.ReadFile(it.Path)
	if err != nil {
		fmt.Printf("❌ [Processor] 입력 파일 읽기 실패 (%s): %v\n", it.Path, err)
		return fmt.Errorf("read input file (%s): %w", it.Path, err)
	}
	fmt.Printf("📄 [Processor] 입력 파일 읽기 성공 (%d bytes)\n", len(data))

	// 2) 대문자로 변환
	upper := strings.ToUpper(string(data))
	fmt.Printf("🔠 [Processor] 대문자 변환 완료 (예시: %.30s...)\n", upper)

	// 3) 출력 경로 결정
	outPath := filepath.Join(u.OutDir, fmt.Sprintf("output-%s.txt", it.ID))
	fmt.Printf("📤 [Processor] 출력 경로: %s\n", outPath)

	// 4) 출력 파일 저장
	if err := os.WriteFile(outPath, []byte(upper), 0644); err != nil {
		fmt.Printf("❌ [Processor] 출력 파일 저장 실패: %v\n", err)
		return fmt.Errorf("write output file (%s): %w", outPath, err)
	}

	fmt.Printf("✅ [Processor] 처리 완료: %s → %s\n", it.Path, outPath)
	return nil
}
