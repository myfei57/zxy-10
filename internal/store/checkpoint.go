package store

import "time"

// Checkpoint is the durable read cursor of one log source. It must only
// advance after the corresponding batch data is durably stored, so recovery
// can resume exactly where the pipeline left off.
type Checkpoint struct {
	SourceID  string    `json:"source_id"`
	Offset    int64     `json:"offset"`
	Line      int       `json:"line"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CheckpointStore persists per-source read cursors.
type CheckpointStore struct {
	fs *FileStore
}

// NewCheckpointStore creates the checkpoint store.
func NewCheckpointStore(fs *FileStore) *CheckpointStore {
	return &CheckpointStore{fs: fs}
}

// Path resolves the checkpoint file of a source.
func (c *CheckpointStore) Path(sourceID string) string {
	return c.fs.Path("checkpoints", sourceID)
}

// Save durably writes the checkpoint of a source.
func (c *CheckpointStore) Save(checkpoint Checkpoint) error {
	return c.fs.WriteJSON(c.Path(checkpoint.SourceID), checkpoint)
}

// Load reads the checkpoint of a source, mapping a missing record to
// ErrNotFound.
func (c *CheckpointStore) Load(sourceID string) (Checkpoint, error) {
	var checkpoint Checkpoint
	if err := c.fs.ReadJSON(c.Path(sourceID), &checkpoint); err != nil {
		return Checkpoint{}, err
	}
	return checkpoint, nil
}
