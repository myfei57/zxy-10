package quota

import "time"

// Recover restores a source quota after its window resets and clears the
// flow-control limiter so the source resumes full rate.
func (s *Service) Recover(sourceID string) {
	bucket := s.bucket(sourceID)
	bucket.Tokens = bucket.Capacity
	bucket.LastRefill = time.Now().UTC()
}
