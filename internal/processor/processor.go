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

// 샘플: 파일 내용을 읽고 대문자로 변환 후 저장
type UppercaseFile struct{ OutDir string }

func (u UppercaseFile) Process(ctx context.Context, it Item) error {
	// 1) 입력 파일 읽기
	data, err := os.ReadFile(it.Path)
	if err != nil {
		return fmt.Errorf("read input file (%s): %w", it.Path, err)
	}

	// 2) 대문자로 변환
	upper := strings.ToUpper(string(data))

	// 3) 출력 경로 결정 (OutDir/output-ID.txt)
	outPath := filepath.Join(u.OutDir, fmt.Sprintf("output-%s.txt", it.ID))

	// 4) 출력 파일 저장
	if err := os.WriteFile(outPath, []byte(upper), 0644); err != nil {
		return fmt.Errorf("write output file (%s): %w", outPath, err)
	}

	fmt.Printf("[Processor] processed %s → %s\n", it.Path, outPath)

	return nil
}
