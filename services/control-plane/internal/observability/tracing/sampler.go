package tracing

import (
	"encoding/binary"
	"hash/fnv"
)

type DecisionInput struct {
	TenantID   string
	RequestID  string
	StatusCode int
	DurationMS int64
}

type Sampler struct {
	defaultRate float64
	slowMS      int64
}

func NewSampler(defaultRate float64, slowRequestMS int64) *Sampler {
	if defaultRate < 0 {
		defaultRate = 0
	}

	if defaultRate > 1 {
		defaultRate = 1
	}

	return &Sampler{
		defaultRate: defaultRate,
		slowMS:      slowRequestMS,
	}
}

func (s *Sampler) ShouldSample(in DecisionInput) bool {
	if in.StatusCode >= 500 || in.DurationMS >= s.slowMS {
		return true
	}

	if s.defaultRate >= 1 {
		return true
	}

	if s.defaultRate <= 0 {
		return false
	}

	h := fnv.New64a()
	_, _ = h.Write([]byte(in.TenantID))
	_, _ = h.Write([]byte(":"))
	_, _ = h.Write([]byte(in.RequestID))
	sum := h.Sum64()

	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], sum)
	u := float64(binary.BigEndian.Uint64(buf[:])%1000000) / 1000000.0

	return u < s.defaultRate
}
