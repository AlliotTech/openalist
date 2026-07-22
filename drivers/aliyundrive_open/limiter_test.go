package aliyundrive_open

import (
	"context"
	"errors"
	"testing"
)

func resetLimitersForTest() {
	limitersLock.Lock()
	defer limitersLock.Unlock()
	limiters = make(map[string]*limiter)
}

func TestLimiterIsSharedByUser(t *testing.T) {
	resetLimitersForTest()
	t.Cleanup(resetLimitersForTest)

	first := getLimiterForUser("user-1")
	second := getLimiterForUser("user-1")
	other := getLimiterForUser("user-2")

	if first != second {
		t.Fatal("same user should share one limiter")
	}
	if first == other {
		t.Fatal("different users should not share a limiter")
	}
	if first.usedBy != 2 {
		t.Fatalf("shared limiter usedBy = %d, want 2", first.usedBy)
	}

	first.free()
	if _, ok := limiters["user-1"]; !ok {
		t.Fatal("limiter was removed while still in use")
	}
	second.free()
	if _, ok := limiters["user-1"]; ok {
		t.Fatal("unused limiter was not removed")
	}
	other.free()
}

func TestGlobalLimiterIsRetained(t *testing.T) {
	resetLimitersForTest()
	t.Cleanup(resetLimitersForTest)

	lim := getLimiterForUser(globalLimiterUserID)
	lim.free()
	if limiters[globalLimiterUserID] != lim {
		t.Fatal("global limiter should be retained for initialization requests")
	}
}

func TestReferenceUsesSourceLimiter(t *testing.T) {
	resetLimitersForTest()
	t.Cleanup(resetLimitersForTest)

	source := &AliyundriveOpen{limiter: getLimiterForUser("user-1")}
	reference := &AliyundriveOpen{ref: source}

	if err := reference.wait(context.Background(), limiterList); err != nil {
		t.Fatalf("reference wait failed: %v", err)
	}
	source.limiter.free()
}

func TestLimiterWaitHonorsCanceledContext(t *testing.T) {
	resetLimitersForTest()
	t.Cleanup(resetLimitersForTest)

	lim := getLimiterForUser("user-1")
	defer lim.free()
	if err := lim.wait(context.Background(), limiterLink); err != nil {
		t.Fatalf("initial wait failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := lim.wait(ctx, limiterLink); !errors.Is(err, context.Canceled) {
		t.Fatalf("wait error = %v, want context canceled", err)
	}
}
