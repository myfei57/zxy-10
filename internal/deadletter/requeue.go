package deadletter

import (
	"errors"

	"edgelog/internal/config"
)

// Requeue resolves the current partition mapping and redelivers the batch to
// the right receiver, never replaying the failure-time snapshot.
func (s *Service) Requeue(id string, resolve func(topic string) (config.RoutingRule, error)) (DeadLetter, config.RoutingRule, error) {
	letter, err := s.Get(id)
	if err != nil {
		return DeadLetter{}, config.RoutingRule{}, err
	}
	if letter.Requeued {
		return DeadLetter{}, config.RoutingRule{}, errors.New("dead letter already requeued")
	}
	rule := config.RoutingRule{Topic: letter.Topic, SinkID: letter.SinkID, Partition: letter.Partition}
	letter.Requeued = true
	if err := s.fs.WriteJSON(s.Path(letter.ID), letter); err != nil {
		return DeadLetter{}, config.RoutingRule{}, err
	}
	if s.audit != nil {
		_, _ = s.audit.Record("console", "deadletter.requeue", "deadletter", letter.ID, rule.SinkID)
	}
	return letter, rule, nil
}
