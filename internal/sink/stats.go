package sink

// Summary is the aggregate sink state for the console.
type Summary struct {
	Count       int `json:"count"`
	Delivered   int `json:"delivered"`
	Armed       int `json:"armed"`
	TotalFailed int `json:"total_failed"`
}

// Summary aggregates every sink in the registry.
func (r *Registry) Summary() (Summary, error) {
	sinks, err := r.List()
	if err != nil {
		return Summary{}, err
	}
	summary := Summary{Count: len(sinks)}
	for _, sink := range sinks {
		summary.Delivered += sink.Delivered
		if sink.FailuresRemaining > 0 {
			summary.Armed++
		}
		summary.TotalFailed += sink.FailuresRemaining
	}
	return summary, nil
}
