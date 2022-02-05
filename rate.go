package renda

import (
	"math"
	"time"
)

type Rate struct {
	Freq int
	Per  time.Duration
}

func (cp Rate) Pace(elapsed time.Duration, hits uint64) (time.Duration, bool) {
	switch {
	case cp.Per == 0 || cp.Freq == 0:
		return 0, false
	case cp.Per < 0 || cp.Freq < 0:
		return 0, true
	}

	expectedHits := uint64(cp.Freq) * uint64(elapsed/cp.Per)
	if hits < expectedHits {
		return 0, false
	}
	interval := uint64(cp.Per.Nanoseconds() / int64(cp.Freq))
	if math.MaxInt64/interval < hits {
		return 0, true
	}
	delta := time.Duration((hits + 1) * interval)
	return delta - elapsed, false
}
