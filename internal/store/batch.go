package store

import "time"

// BatchState tracks the lifecycle of one staged log batch.
type BatchState string

const (
	BatchAssembled BatchState = "assembled"
	BatchStaged    BatchState = "staged"
	BatchCommitted BatchState = "committed"
	BatchFailed    BatchState = "failed"
	BatchDead      BatchState = "dead"
)

// BatchRecord is the durable record of one log batch.
type BatchRecord struct {
	ID        string     `json:"id"`
	SourceID  string     `json:"source_id"`
	Topic     string     `json:"topic"`
	Lines     []string   `json:"lines"`
	EndOffset int64      `json:"end_offset"`
	State     BatchState `json:"state"`
	Retries   int        `json:"retries"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// BatchStore persists staged batches and their delivery state.
type BatchStore struct {
	fs *FileStore
}

// NewBatchStore creates the batch store.
func NewBatchStore(fs *FileStore) *BatchStore {
	return &BatchStore{fs: fs}
}

// Path resolves the record of a batch.
func (b *BatchStore) Path(id string) string {
	return b.fs.Path("batches", id)
}

// Stage durably writes a batch record.
func (b *BatchStore) Stage(record BatchRecord) error {
	return b.fs.WriteJSON(b.Path(record.ID), record)
}

// Load reads one batch record.
func (b *BatchStore) Load(id string) (BatchRecord, error) {
	var record BatchRecord
	if err := b.fs.ReadJSON(b.Path(id), &record); err != nil {
		return BatchRecord{}, err
	}
	return record, nil
}

// List returns every batch record id, newest first.
func (b *BatchStore) List() ([]string, error) {
	return b.fs.List("batches")
}
