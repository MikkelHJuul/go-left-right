package primitive

import "time"

type config struct {
	maxWrites int
	ticker    *time.Ticker
}

func (c config) Close() error {
	c.ticker.Stop()
	return nil
}

func WithMaxSyncDuration(duration time.Duration) func(*config) {
	return func(cfg *config) {
		cfg.ticker = time.NewTicker(duration)
	}
}

func WithMaxNumWritesPerSync(num int) func(*config) {
	return func(cfg *config) {
		cfg.maxWrites = num
	}
}
