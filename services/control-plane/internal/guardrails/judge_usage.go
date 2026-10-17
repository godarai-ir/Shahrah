package guardrails

import (
	"context"
	"strings"
	"sync"
)

type judgeUsageKey struct{}

// JudgeUsage counts the semantic-judge calls one request made, so the caller
// can charge them. Answers served from the judge cache are not counted.
type JudgeUsage struct {
	mu    sync.Mutex
	calls int
	model string
}

// WithJudgeUsage returns a context whose judge calls are counted in usage.
func WithJudgeUsage(ctx context.Context, usage *JudgeUsage) context.Context {
	return context.WithValue(ctx, judgeUsageKey{}, usage)
}

// Calls returns the number of judge calls that reached the judge provider.
func (u *JudgeUsage) Calls() int {
	if u == nil {
		return 0
	}

	u.mu.Lock()
	defer u.mu.Unlock()

	return u.calls
}

// Model returns the judge model of the last counted call.
func (u *JudgeUsage) Model() string {
	if u == nil {
		return ""
	}

	u.mu.Lock()
	defer u.mu.Unlock()

	return u.model
}

// Reset clears the count after the caller charged it.
func (u *JudgeUsage) Reset() {
	if u == nil {
		return
	}

	u.mu.Lock()
	defer u.mu.Unlock()

	u.calls = 0
}

// recordJudgeCall counts a successful, uncached judge call on the context.
func recordJudgeCall(ctx context.Context, response JudgeResponse, err error) {
	usage, ok := ctx.Value(judgeUsageKey{}).(*JudgeUsage)
	if !ok || usage == nil || err != nil || response.Cached {
		return
	}

	usage.mu.Lock()
	defer usage.mu.Unlock()

	usage.calls++

	if model := strings.TrimSpace(response.Model); model != "" {
		usage.model = model
	}
}
