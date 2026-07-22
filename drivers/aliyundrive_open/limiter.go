package aliyundrive_open

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/time/rate"
)

// API rate limits are per user and application, so storage instances using
// the same account must share one limiter.
type limiterType int

const (
	limiterList limiterType = iota
	limiterLink
	limiterOther
)

const (
	listRateLimit       = 3.9
	linkRateLimit       = 0.9
	otherRateLimit      = 14.9
	globalLimiterUserID = ""
)

type limiter struct {
	usedBy int
	list   *rate.Limiter
	link   *rate.Limiter
	other  *rate.Limiter
}

var (
	limiters     = make(map[string]*limiter)
	limitersLock sync.Mutex
)

func getLimiterForUser(userID string) *limiter {
	limitersLock.Lock()
	defer limitersLock.Unlock()
	defer cleanupLimitersLocked()

	if lim, ok := limiters[userID]; ok {
		lim.usedBy++
		return lim
	}
	lim := &limiter{
		usedBy: 1,
		list:   rate.NewLimiter(rate.Limit(listRateLimit), 1),
		link:   rate.NewLimiter(rate.Limit(linkRateLimit), 1),
		other:  rate.NewLimiter(rate.Limit(otherRateLimit), 1),
	}
	limiters[userID] = lim
	return lim
}

func cleanupLimitersLocked() {
	for id, lim := range limiters {
		if lim.usedBy <= 0 && id != globalLimiterUserID {
			delete(limiters, id)
		}
	}
}

func (l *limiter) wait(ctx context.Context, typ limiterType) error {
	if l == nil {
		return fmt.Errorf("driver not init")
	}
	switch typ {
	case limiterList:
		return l.list.Wait(ctx)
	case limiterLink:
		return l.link.Wait(ctx)
	case limiterOther:
		return l.other.Wait(ctx)
	default:
		return fmt.Errorf("unknown limiter type: %d", typ)
	}
}

func (l *limiter) free() {
	if l == nil {
		return
	}
	limitersLock.Lock()
	defer limitersLock.Unlock()
	if l.usedBy > 0 {
		l.usedBy--
	}
	cleanupLimitersLocked()
}

func (d *AliyundriveOpen) wait(ctx context.Context, typ limiterType) error {
	if d == nil {
		return fmt.Errorf("driver not init")
	}
	if d.ref != nil {
		return d.ref.wait(ctx, typ)
	}
	return d.limiter.wait(ctx, typ)
}
