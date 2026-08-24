package audit

import (
	"time"

	"github.com/google/uuid"

	"edgelog/internal/store"
)

// Entry is one persisted operation-audit record.
type Entry struct {
	ID         string `json:"id"`
	At         string `json:"at"`
	Actor      string `json:"actor"`
	Action     string `json:"action"`
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	Detail     string `json:"detail"`
}

// Recorder is the audit sink implemented by Service and injected into the
// domain components that mutate control state.
type Recorder interface {
	Record(actor, action, targetType, targetID, detail string) (Entry, error)
}

// Service appends audit entries to the file store.
type Service struct {
	fs    *store.FileStore
	clock func() time.Time
}

// NewService creates the audit service.
func NewService(fs *store.FileStore) *Service {
	return &Service{fs: fs, clock: time.Now}
}

// Record persists one audit entry.
func (s *Service) Record(actor, action, targetType, targetID, detail string) (Entry, error) {
	entry := Entry{
		ID:         uuid.NewString(),
		At:         s.clock().UTC().Format(time.RFC3339),
		Actor:      actor,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Detail:     detail,
	}
	if err := s.fs.WriteJSON(s.fs.Path("audit", entry.ID), entry); err != nil {
		return Entry{}, err
	}
	return entry, nil
}
