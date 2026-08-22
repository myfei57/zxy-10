package router

import (
	"time"

	"github.com/google/uuid"

	"edgelog/internal/config"
	"edgelog/internal/store"
)

// Subscription binds a topic to a sink at a rules version.
type Subscription struct {
	ID           string `json:"id"`
	Topic        string `json:"topic"`
	SinkID       string `json:"sink_id"`
	RulesVersion int    `json:"rules_version"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// SubscriptionService persists topic subscriptions.
type SubscriptionService struct {
	fs    *store.FileStore
	rules *config.RulesService
	clock func() time.Time
}

// NewSubscriptionService creates the subscription service.
func NewSubscriptionService(fs *store.FileStore, rules *config.RulesService) *SubscriptionService {
	return &SubscriptionService{fs: fs, rules: rules, clock: time.Now}
}

func (s *SubscriptionService) now() string {
	return s.clock().UTC().Format(time.RFC3339)
}

// Path resolves the subscription record.
func (s *SubscriptionService) Path(id string) string {
	return s.fs.Path("subscriptions", id)
}

// Bind creates or updates a subscription for a topic at the current rules
// version.
func (s *SubscriptionService) Bind(topic, sinkID string) (Subscription, error) {
	ruleset, err := s.rules.Current()
	if err != nil {
		return Subscription{}, err
	}
	sub := Subscription{
		ID:           uuid.NewString(),
		Topic:        topic,
		SinkID:       sinkID,
		RulesVersion: ruleset.Version,
		CreatedAt:    s.now(),
		UpdatedAt:    s.now(),
	}
	if err := s.fs.WriteJSON(s.Path(sub.ID), sub); err != nil {
		return Subscription{}, err
	}
	return sub, nil
}

// List returns every subscription.
func (s *SubscriptionService) List() ([]Subscription, error) {
	ids, err := s.fs.List("subscriptions")
	if err != nil {
		return nil, err
	}
	out := make([]Subscription, 0, len(ids))
	for _, id := range ids {
		var sub Subscription
		if err := s.fs.ReadJSON(s.Path(id), &sub); err != nil {
			return nil, err
		}
		out = append(out, sub)
	}
	return out, nil
}

// Rebind refreshes every subscription's sink against the current durable
// ruleset.
func (s *SubscriptionService) Rebind() error {
	ruleset, err := s.rules.Current()
	if err != nil {
		return err
	}
	subscriptions, err := s.List()
	if err != nil {
		return err
	}
	for _, sub := range subscriptions {
		rule, err := findRule(ruleset, sub.Topic)
		if err != nil {
			continue
		}
		sub.SinkID = rule.SinkID
		sub.RulesVersion = ruleset.Version
		sub.UpdatedAt = s.now()
		if err := s.fs.WriteJSON(s.Path(sub.ID), sub); err != nil {
			return err
		}
	}
	return nil
}

func findRule(ruleset config.Ruleset, topic string) (config.RoutingRule, error) {
	for _, rule := range ruleset.Rules {
		if rule.Topic == topic {
			return rule, nil
		}
	}
	return config.RoutingRule{}, errNoRule
}
