package tailer

import (
	"fmt"

	"edgelog/internal/store"
)

// Cursor is the in-memory read position of one source.
type Cursor struct {
	SourceID string `json:"source_id"`
	Offset   int64  `json:"offset"`
	Line     int    `json:"line"`
}

// Current loads the persisted checkpoint of a source, or starts at zero.
func (s *Service) Current(sourceID string) (Cursor, error) {
	if cached, ok := s.current[sourceID]; ok {
		return cached, nil
	}
	checkpoint, err := s.checkpoints.Load(sourceID)
	if err != nil {
		if err == store.ErrNotFound {
			cursor := Cursor{SourceID: sourceID}
			s.current[sourceID] = cursor
			return cursor, nil
		}
		return Cursor{}, err
	}
	cursor := Cursor{SourceID: sourceID, Offset: checkpoint.Offset, Line: checkpoint.Line}
	s.current[sourceID] = cursor
	return cursor, nil
}

// Advance persists the new cursor and only then updates the in-memory
// position. A failed checkpoint write must never leave the cursor ahead of
// the durable state.
func (s *Service) Advance(current Cursor, next Cursor) (Cursor, error) {
	if current.SourceID != next.SourceID {
		return Cursor{}, fmt.Errorf("cursor source mismatch: %s -> %s", current.SourceID, next.SourceID)
	}
	checkpoint := store.Checkpoint{
		SourceID:  next.SourceID,
		Offset:    next.Offset,
		Line:      next.Line,
		UpdatedAt: s.now(),
	}
	s.current[next.SourceID] = next
	if err := s.checkpoints.Save(checkpoint); err != nil {
		return Cursor{}, err
	}
	return next, nil
}

// Reset rewinds a source cursor to the start after the underlying log file
// rotated.
func (s *Service) Reset(sourceID string) (Cursor, error) {
	checkpoint := store.Checkpoint{
		SourceID:  sourceID,
		Offset:    0,
		Line:      0,
		UpdatedAt: s.now(),
	}
	if err := s.checkpoints.Save(checkpoint); err != nil {
		return Cursor{}, err
	}
	cursor := Cursor{SourceID: sourceID}
	s.current[sourceID] = cursor
	return cursor, nil
}

// String renders a compact cursor summary.
func (c Cursor) String() string {
	return fmt.Sprintf("%s@%d:%d", c.SourceID, c.Offset, c.Line)
}
