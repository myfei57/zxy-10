package router

import "edgelog/internal/config"

// PartitionForTopic resolves the current partition of a topic through the
// live routing rules.
func (r *Router) PartitionForTopic(topic string) (config.RoutingRule, error) {
	return r.rules.PartitionFor(topic)
}
