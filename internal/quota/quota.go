package quota

import (
	"time"
)

// TokenBucket refills over a window and allows one token per consumed line.
type TokenBucket struct {
	Tokens     float64   `json:"tokens"`
	Capacity   float64   `json:"capacity"`
	LastRefill time.Time `json:"last_refill"`
}

// LimiterResetter is implemented by the router flow controller.
type LimiterResetter interface {
	ResetLimiter(sourceID string)
}

// Service tracks per-source quota windows. When a window is exhausted the
// caller throttles the source; recovery must reset the limiter.
type Service struct {
	buckets       map[string]*TokenBucket
	capacity      float64
	windowSeconds int
	limiter       LimiterResetter
	clock         func() time.Time
}

// NewService creates the quota service.
func NewService(capacity float64, windowSeconds int) *Service {
	return &Service{
		buckets:       make(map[string]*TokenBucket),
		capacity:      capacity,
		windowSeconds: windowSeconds,
		clock:         time.Now,
	}
}

// NewServiceWithLimiter wires the flow limiter into the quota service.
func NewServiceWithLimiter(capacity float64, windowSeconds int, limiter LimiterResetter) *Service {
	s := NewService(capacity, windowSeconds)
	s.limiter = limiter
	return s
}

func (s *Service) bucket(sourceID string) *TokenBucket {
	bucket, ok := s.buckets[sourceID]
	if !ok {
		bucket = &TokenBucket{
			Tokens:     s.capacity,
			Capacity:   s.capacity,
			LastRefill: s.clock().UTC(),
		}
		s.buckets[sourceID] = bucket
	}
	return bucket
}

// Allow consumes one token, refilling by the elapsed window first. It returns
// false when the bucket is empty.
func (s *Service) Allow(sourceID string) bool {
	bucket := s.bucket(sourceID)
	now := s.clock().UTC()
	elapsed := now.Sub(bucket.LastRefill).Seconds()
	if elapsed >= float64(s.windowSeconds) {
		bucket.Tokens = bucket.Capacity
		bucket.LastRefill = now
	}
	if bucket.Tokens >= 1 {
		bucket.Tokens--
		return true
	}
	return false
}
