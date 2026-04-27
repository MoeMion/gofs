package ftpsync

import (
	"context"
	"fmt"
	"time"
)

func retryWithContext(ctx context.Context, opts RetryOptions, operation string, fn func() error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	attempts := opts.Count
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := fn(); err != nil {
			lastErr = err
		} else {
			return nil
		}
		if attempt == attempts || opts.Wait <= 0 {
			continue
		}
		timer := time.NewTimer(opts.Wait)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return ctx.Err()
		case <-timer.C:
		}
	}
	if operation == "" {
		return lastErr
	}
	return fmt.Errorf("%s failed after %d attempt(s): %w", operation, attempts, lastErr)
}
