package mode

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tryoo0607/job-test/internal/config"
	"github.com/tryoo0607/job-test/internal/processor"
	"github.com/tryoo0607/job-test/internal/runtime"
)

func RunFixed(ctx context.Context, cfg config.Config) error {
	fmt.Println("🔧 [RunFixed] 시작")

	items := loadItems(cfg.Items, cfg.ItemsFile, cfg.InputDir)
	fmt.Printf("🔧 [RunFixed] 총 로드된 items 개수: %d\n", len(items))
	for _, it := range items {
		fmt.Printf("    ↳ ID: %s, Path: %s\n", it.ID, it.Path)
	}

	ch := make(chan processor.Item)

	go func() {
		defer close(ch)
		for _, it := range items {
			fmt.Printf("📤 [RunFixed] 채널에 아이템 전송: ID=%s\n", it.ID)
			ch <- it
		}
	}()

	p := processor.UppercaseFile{OutDir: cfg.OutputDir}
	fmt.Printf("🔧 [RunFixed] Processor 생성 완료. OutputDir=%s\n", cfg.OutputDir)

	err := runtime.RunPool(ctx, cfg.MaxConcurrency, ch, p)
	if err != nil {
		fmt.Printf("❌ [RunFixed] RunPool 실패: %v\n", err)
	} else {
		fmt.Println("✅ [RunFixed] RunPool 성공")
	}
	return err
}

func loadItems(values []string, itemList string, inputDir string) []processor.Item {
	var items []processor.Item

	fmt.Printf("🔍 [loadItems] values 개수: %d\n", len(values))
	for i, v := range values {
		// 절대 경로 또는 상대 경로로 직접 지정된 경우
		fmt.Printf("📥 [loadItems] values 추가: val-%d = %s\n", i, v)
		items = append(items, processor.Item{
			ID:   fmt.Sprintf("val-%d", i),
			Path: v,
		})
	}

	if itemList != "" {
		fmt.Printf("📂 [loadItems] items-file 경로: %s\n", itemList)
		f, err := os.Open(itemList)
		if err != nil {
			fmt.Printf("❌ [loadItems] 파일 열기 실패: %v\n", err)
		} else {
			defer f.Close()
			scanner := bufio.NewScanner(f)
			idx := len(items)
			for scanner.Scan() {
				rel := scanner.Text()
				abs := filepath.Join(inputDir, rel)
				fmt.Printf("📥 [loadItems] file 추가: file-%d = %s (⇒ %s)\n", idx, rel, abs)
				items = append(items, processor.Item{
					ID:   fmt.Sprintf("file-%d", idx),
					Path: abs,
				})
				idx++
			}
			if err := scanner.Err(); err != nil {
				fmt.Printf("❌ [loadItems] 파일 스캔 중 오류: %v\n", err)
			}
		}
	} else {
		fmt.Println("ℹ️ [loadItems] filePath가 비어 있음 → 파일로부터 추가 없음")
	}

	return items
}
