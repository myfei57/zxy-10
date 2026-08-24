package deadletter

import (
	"time"

	"github.com/google/uuid"

	"edgelog/internal/audit"
	"edgelog/internal/store"
)

// DeadLetter is a batch held after its retry budget was exhausted.
type DeadLetter struct {
	ID        string    `json:"id"`
	BatchID   string    `json:"batch_id"`
	SourceID  string    `json:"source_id"`
	Topic     string    `json:"topic"`
	SinkID    string    `json:"sink_id"`
	Partition int       `json:"partition"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
	Requeued  bool      `json:"requeued"`
}

// Service holds and requeues dead letters.
type Service struct {
	fs    *store.FileStore
	audit audit.Recorder
	clock func() time.Time
}

// NewService creates the dead-letter service.
func NewService(fs *store.FileStore) *Service {
	return &Service{fs: fs, clock: time.Now}
}

// NewServiceWithAudit wires the audit sink into the dead-letter service.
func NewServiceWithAudit(fs *store.FileStore, recorder audit.Recorder) *Service {
	s := NewService(fs)
	s.audit = recorder
	return s
}

func (s *Service) now() time.Time {
	return s.clock().UTC()
}

// Path resolves the dead-letter record.
func (s *Service) Path(id string) string {
	return s.fs.Path("deadletters", id)
}

// Hold persists a dead letter for a batch.
func (s *Service) Hold(batchID, sourceID, topic, sinkID string, partition int, reason string) (DeadLetter, error) {
	letter := DeadLetter{
		ID:        uuid.NewString(),
		BatchID:   batchID,
		SourceID:  sourceID,
		Topic:     topic,
		SinkID:    sinkID,
		Partition: partition,
		Reason:    reason,
		CreatedAt: s.now(),
	}
	if err := s.fs.WriteJSON(s.Path(letter.ID), letter); err != nil {
		return DeadLetter{}, err
	}
	if s.audit != nil {
		_, _ = s.audit.Record("engine", "deadletter.hold", "batch", batchID, reason)
	}
	return letter, nil
}

// Get loads one dead letter.
func (s *Service) Get(id string) (DeadLetter, error) {
	var letter DeadLetter
	if err := s.fs.ReadJSON(s.Path(id), &letter); err != nil {
		return DeadLetter{}, err
	}
	return letter, nil
}

// List returns every dead letter, newest first.
func (s *Service) List() ([]DeadLetter, error) {
	ids, err := s.fs.Latest("deadletters", 200)
	if err != nil {
		return nil, err
	}
	out := make([]DeadLetter, 0, len(ids))
	for _, id := range ids {
		letter, err := s.Get(id)
		if err != nil {
			return nil, err
		}
		out = append(out, letter)
	}
	return out, nil
}
