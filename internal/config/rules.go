package config

import (
	"errors"
	"time"

	"edgelog/internal/audit"
	"edgelog/internal/store"
)

// RoutingRule maps one topic to a sink and a partition.
type RoutingRule struct {
	Topic     string `json:"topic"`
	SinkID    string `json:"sink_id"`
	Partition int    `json:"partition"`
}

// Ruleset is the durable, versioned set of routing rules.
type Ruleset struct {
	Version int           `json:"version"`
	Rules   []RoutingRule `json:"rules"`
	Updated string        `json:"updated"`
}

// RulesService persists routing rules and resolves partitions.
type RulesService struct {
	fs       *store.FileStore
	audit    audit.Recorder
	rebinder RulesRebinder
	clock    func() time.Time
}

// RulesRebinder is implemented by the router subscription service so a rules
// update can refresh topic bindings.
type RulesRebinder interface {
	Rebind() error
}

// NewRulesService creates the routing rules service.
func NewRulesService(fs *store.FileStore) *RulesService {
	return &RulesService{fs: fs, clock: time.Now}
}

// NewRulesServiceWithAudit wires the audit sink into the rules service.
func NewRulesServiceWithAudit(fs *store.FileStore, recorder audit.Recorder) *RulesService {
	s := NewRulesService(fs)
	s.audit = recorder
	return s
}

func (s *RulesService) now() string {
	return s.clock().UTC().Format(time.RFC3339)
}

// SetRebinder wires the subscription rebind callback.
func (s *RulesService) SetRebinder(rebinder RulesRebinder) {
	s.rebinder = rebinder
}

// Path resolves the durable ruleset record.
func (s *RulesService) Path() string {
	return s.fs.Path("rules", "current")
}

// Current loads the current ruleset, returning an empty one when absent.
func (s *RulesService) Current() (Ruleset, error) {
	var ruleset Ruleset
	if err := s.fs.ReadJSON(s.Path(), &ruleset); err != nil {
		if err == store.ErrNotFound {
			return Ruleset{Version: 1}, nil
		}
		return Ruleset{}, err
	}
	return ruleset, nil
}

// Update durably replaces the routing rules and bumps the version.
func (s *RulesService) Update(rules []RoutingRule) (Ruleset, error) {
	if len(rules) == 0 {
		return Ruleset{}, errors.New("routing rules cannot be empty")
	}
	current, err := s.Current()
	if err != nil {
		return Ruleset{}, err
	}
	current.Version++
	current.Rules = rules
	current.Updated = s.now()
	if err := s.fs.WriteJSON(s.Path(), current); err != nil {
		return Ruleset{}, err
	}
	if s.rebinder != nil {
		if err := s.rebinder.Rebind(); err != nil {
			return Ruleset{}, err
		}
	}
	if s.audit != nil {
		_, _ = s.audit.Record("console", "rules.update", "rules", "current", current.String())
	}
	return current, nil
}

// PartitionFor resolves the current partition of a topic, or returns an
// error when no rule matches.
func (s *RulesService) PartitionFor(topic string) (RoutingRule, error) {
	ruleset, err := s.Current()
	if err != nil {
		return RoutingRule{}, err
	}
	for _, rule := range ruleset.Rules {
		if rule.Topic == topic {
			return rule, nil
		}
	}
	return RoutingRule{}, errors.New("no routing rule for topic " + topic)
}

// String renders a compact summary of a ruleset for audit entries.
func (r Ruleset) String() string {
	return "v" + itoa(r.Version) + " rules=" + itoa(len(r.Rules))
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	digits := []byte{}
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	return string(digits)
}
