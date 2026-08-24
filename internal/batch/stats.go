package batch

import "edgelog/internal/store"

// Summary aggregates every batch record for the console.
func (s *Service) Summary() (store.BatchSummary, error) {
	records, err := s.List()
	if err != nil {
		return store.BatchSummary{}, err
	}
	return store.SummarizeBatches(records), nil
}

// TopicSummary returns per-topic batch totals for the console.
func (s *Service) TopicSummary() ([]store.TopicSummary, error) {
	records, err := s.List()
	if err != nil {
		return nil, err
	}
	return store.SummarizeByTopic(records), nil
}
