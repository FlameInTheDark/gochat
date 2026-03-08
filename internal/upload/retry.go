package upload

import (
	"context"
	"time"
)

func Retry(ctx context.Context, attempts int, delay time.Duration, fn func(context.Context) error) error {
	if attempts <= 0 {
		attempts = 1
	}

	var err error
	currentDelay := delay
	for i := 0; i < attempts; i++ {
		err = fn(ctx)
		if err == nil {
			return nil
		}
		if i == attempts-1 {
			return err
		}

		timer := time.NewTimer(currentDelay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}

		if currentDelay > 0 && currentDelay < time.Second {
			currentDelay *= 2
		}
	}
	return err
}
