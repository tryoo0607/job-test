package runtime

import (
	"context"
	"fmt"
	"time"
)

func Hold(ctx context.Context, d time.Duration) {
	if d <= 0 {
		return
	}
	fmt.Printf("⏳ holding pod for %s...\n", d)
	select {
	case <-time.After(d):
	case <-ctx.Done():
	}
}
