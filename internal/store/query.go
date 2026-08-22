package store

// BatchSummary aggregates batch records for console reporting.
type BatchSummary struct {
	Total      int                `json:"total"`
	ByState    map[BatchState]int `json:"by_state"`
	TotalLines int                `json:"total_lines"`
}

// SummarizeBatches groups records by state and sums their line counts.
func SummarizeBatches(records []BatchRecord) BatchSummary {
	summary := BatchSummary{Total: len(records), ByState: make(map[BatchState]int)}
	for _, record := range records {
		summary.ByState[record.State]++
		summary.TotalLines += len(record.Lines)
	}
	return summary
}

// TopicSummary aggregates batch and line counts for one topic.
type TopicSummary struct {
	Topic   string `json:"topic"`
	Batches int    `json:"batches"`
	Lines   int    `json:"lines"`
}

// SummarizeByTopic groups records by topic for the console.
func SummarizeByTopic(records []BatchRecord) []TopicSummary {
	order := make([]string, 0)
	byTopic := make(map[string]*TopicSummary)
	for _, record := range records {
		summary, ok := byTopic[record.Topic]
		if !ok {
			summary = &TopicSummary{Topic: record.Topic}
			byTopic[record.Topic] = summary
			order = append(order, record.Topic)
		}
		summary.Batches++
		summary.Lines += len(record.Lines)
	}
	out := make([]TopicSummary, 0, len(order))
	for _, topic := range order {
		out = append(out, *byTopic[topic])
	}
	return out
}
