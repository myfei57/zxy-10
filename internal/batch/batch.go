package batch

import (
	"time"

	"github.com/google/uuid"

	"edgelog/internal/audit"
	"edgelog/internal/store"
	"edgelog/internal/tailer"
)

// Service assembles, stages, commits and results log batches.
type Service struct {
	fs          *store.FileStore
	batches     *store.BatchStore
	checkpoints *store.CheckpointStore
	dedup       DedupWindow
	advancer    CursorAdvancer
	audit       audit.Recorder
	clock       func() time.Time
}

// DedupWindow is the dedup surface used by the commit path.
type DedupWindow interface {
	Accept(sourceID, line string) bool
	Close(sourceID string)
}

// CursorAdvancer is implemented by the tailer service.
type CursorAdvancer interface {
	Current(sourceID string) (tailer.Cursor, error)
	Advance(current, next tailer.Cursor) (tailer.Cursor, error)
}

// NewService creates the batch service.
func NewService(fs *store.FileStore, batches *store.BatchStore, checkpoints *store.CheckpointStore, dedup DedupWindow, advancer CursorAdvancer) *Service {
	return &Service{fs: fs, batches: batches, checkpoints: checkpoints, dedup: dedup, advancer: advancer, clock: time.Now}
}

// NewServiceWithAudit wires the audit sink into the batch service.
func NewServiceWithAudit(fs *store.FileStore, batches *store.BatchStore, checkpoints *store.CheckpointStore, dedup DedupWindow, advancer CursorAdvancer, recorder audit.Recorder) *Service {
	s := NewService(fs, batches, checkpoints, dedup, advancer)
	s.audit = recorder
	return s
}

func (s *Service) now() time.Time {
	return s.clock().UTC()
}

// Load reads one batch record.
func (s *Service) Load(id string) (store.BatchRecord, error) {
	return s.batches.Load(id)
}

// List returns every batch record, newest first.
func (s *Service) List() ([]store.BatchRecord, error) {
	ids, err := s.batches.List()
	if err != nil {
		return nil, err
	}
	out := make([]store.BatchRecord, 0, len(ids))
	for _, id := range ids {
		record, err := s.batches.Load(id)
		if err != nil {
			return nil, err
		}
		out = append(out, record)
	}
	return out, nil
}

// NewID allocates a batch id; exported so tests can keep ids stable.
func NewID() string {
	return uuid.NewString()
}
