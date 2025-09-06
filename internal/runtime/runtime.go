package runtime

import (
	"context"
	"time"
)

func WithRetry(ctx context.Context, max int, initial time.Duration, fn func() error) error {
	backoff := initial

	for attempt := 0; attempt < max; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		// 마지막 시도라면 에러 반환
		if attempt == max-1 {
			return err
		}

		// 백오프 대기 (컨텍스트 취소 고려)
		select {
		case <-time.After(backoff):
			// 다음 시도
			backoff *= 2 // 지수 증가
			if backoff > 30*time.Second {
				backoff = 30 * time.Second // 백오프 상한선
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil // 이 위치에 올 일은 없음 (방어적 코드)
}
