package tailer

import (
	"time"

	"github.com/google/uuid"

	"edgelog/internal/store"
)

// Source is one registered log source.
type Source struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

// Service registers log sources and reads their lines.
type Service struct {
	fs          *store.FileStore
	checkpoints *store.CheckpointStore
	current     map[string]Cursor
	clock       func() time.Time
}

// NewService creates the tailer service.
func NewService(fs *store.FileStore, checkpoints *store.CheckpointStore) *Service {
	return &Service{fs: fs, checkpoints: checkpoints, current: make(map[string]Cursor), clock: time.Now}
}

// Path resolves the source record.
func (s *Service) Path(id string) string {
	return s.fs.Path("sources", id)
}

// Register creates a log source pointing at a local file.
func (s *Service) Register(name, path string) (Source, error) {
	source := Source{ID: uuid.NewString(), Name: name, Path: path}
	if err := s.fs.WriteJSON(s.Path(source.ID), source); err != nil {
		return Source{}, err
	}
	return source, nil
}

// Get loads one source record.
func (s *Service) Get(id string) (Source, error) {
	var source Source
	if err := s.fs.ReadJSON(s.Path(id), &source); err != nil {
		return Source{}, err
	}
	return source, nil
}

// List returns every registered source.
func (s *Service) List() ([]Source, error) {
	ids, err := s.fs.List("sources")
	if err != nil {
		return nil, err
	}
	out := make([]Source, 0, len(ids))
	for _, id := range ids {
		source, err := s.Get(id)
		if err != nil {
			return nil, err
		}
		out = append(out, source)
	}
	return out, nil
}

// now returns the current UTC time for cursor records.
func (s *Service) now() time.Time {
	return s.clock().UTC()
}
