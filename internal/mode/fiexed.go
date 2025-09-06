package mode

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/tryoo0607/job-test/internal/config"
	"github.com/tryoo0607/job-test/internal/processor"
	"github.com/tryoo0607/job-test/internal/runtime"
)

func RunFixed(ctx context.Context, cfg config.Config) error {

	items := loadItems(cfg.Items, cfg.ItemsFile) // []Item

	ch := make(chan processor.Item)

	go func() {
		defer close(ch)
		for _, it := range items {
			ch <- it
		}
	}()

	p := processor.UppercaseFile{OutDir: cfg.OutputDir}

	return runtime.RunPool(ctx, cfg.MaxConcurrency, ch, p)
}

// Indexed는 본인 Index 하나만 처리,
// Fixed는 그거랑 상관없이 Item들을 처리해야 함
func loadItems(values []string, filePath string) []processor.Item {
	var items []processor.Item

	// 1) values에서 바로 추가
	for i, v := range values {
		items = append(items, processor.Item{
			ID:   fmt.Sprintf("val-%d", i),
			Path: v,
		})
	}

	// 2) 파일에서 읽어서 추가
	if filePath != "" {
		f, err := os.Open(filePath)
		if err == nil {
			defer f.Close()
			scanner := bufio.NewScanner(f)
			idx := len(items)
			for scanner.Scan() {
				items = append(items, processor.Item{
					ID:   fmt.Sprintf("file-%d", idx),
					Path: scanner.Text(),
				})
				idx++
			}
		}
	}

	return items
}
