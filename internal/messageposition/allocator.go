package messageposition

import (
	"context"
	"time"

	"github.com/FlameInTheDark/gochat/internal/cache"
)

type channelStore interface {
	GetChannelMessagePosition(ctx context.Context, id int64) (int64, error)
	ReserveMessagePositions(ctx context.Context, id, count int64) (int64, error)
}

func Next(ctx context.Context, c cache.Cache, store channelStore, channelID int64) (int64, error) {
	if store == nil {
		return 0, nil
	}
	if c == nil {
		return store.ReserveMessagePositions(ctx, channelID, 1)
	}

	for {
		current, currentOK := getCachedInt64(ctx, c, CurrentKey(channelID))
		reservedMax, reservedOK := getCachedInt64(ctx, c, ReservedMaxKey(channelID))
		if currentOK && reservedOK && current < reservedMax {
			next, err := c.Incr(ctx, CurrentKey(channelID))
			if err == nil {
				_ = c.SetTTL(ctx, CurrentKey(channelID), CacheTTLSeconds)
				_ = c.SetTTL(ctx, ReservedMaxKey(channelID), CacheTTLSeconds)
				if next <= reservedMax {
					return next, nil
				}
			}
		}

		acquired, err := c.SetTimedJSONNX(ctx, LockKey(channelID), map[string]int64{"channel_id": channelID}, LockTTLSeconds)
		if err != nil {
			return 0, err
		}
		if !acquired {
			select {
			case <-ctx.Done():
				return 0, ctx.Err()
			case <-time.After(10 * time.Millisecond):
				continue
			}
		}

		err = refillReservationUnderLock(ctx, c, store, channelID)
		_ = c.Delete(ctx, LockKey(channelID))
		if err != nil {
			return 0, err
		}
	}
}

func refillReservationUnderLock(ctx context.Context, c cache.Cache, store channelStore, channelID int64) error {
	current, currentOK := getCachedInt64(ctx, c, CurrentKey(channelID))
	if !currentOK {
		dbCursor, err := store.GetChannelMessagePosition(ctx, channelID)
		if err != nil {
			return err
		}
		current = dbCursor
		if err := c.SetTimedInt64(ctx, CurrentKey(channelID), current, CacheTTLSeconds); err != nil {
			return err
		}
	}

	reservedMax, reservedOK := getCachedInt64(ctx, c, ReservedMaxKey(channelID))
	if !reservedOK {
		reservedMax = current
	}

	if current >= reservedMax {
		newReservedMax, err := store.ReserveMessagePositions(ctx, channelID, BlockSize)
		if err != nil {
			return err
		}
		reservedMax = newReservedMax
	}

	if err := c.SetTimedInt64(ctx, ReservedMaxKey(channelID), reservedMax, CacheTTLSeconds); err != nil {
		return err
	}
	return c.SetTTL(ctx, CurrentKey(channelID), CacheTTLSeconds)
}

func getCachedInt64(ctx context.Context, c cache.Cache, key string) (int64, bool) {
	value, err := c.GetInt64(ctx, key)
	if err != nil {
		return 0, false
	}
	return value, true
}
