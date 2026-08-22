package sink

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"edgelog/internal/store"
)

// Sink is one log receiver. Deliver simulates the receiver: while
// FailuresRemaining is positive every delivery fails, which lets operators
// and tests exercise retry and dead-letter paths deterministically.
type Sink struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	FailuresRemaining int       `json:"failures_remaining"`
	Delivered         int       `json:"delivered"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// Registry persists and serves log sinks.
type Registry struct {
	fs *store.FileStore
}

// NewRegistry creates the sink registry.
func NewRegistry(fs *store.FileStore) *Registry {
	return &Registry{fs: fs}
}

// Path resolves the sink record.
func (r *Registry) Path(id string) string {
	return r.fs.Path("sinks", id)
}

// Register creates a sink.
func (r *Registry) Register(name string) (Sink, error) {
	sink := Sink{ID: uuid.NewString(), Name: name, UpdatedAt: time.Now().UTC()}
	if err := r.fs.WriteJSON(r.Path(sink.ID), sink); err != nil {
		return Sink{}, err
	}
	return sink, nil
}

// Get loads one sink.
func (r *Registry) Get(id string) (Sink, error) {
	var sink Sink
	if err := r.fs.ReadJSON(r.Path(id), &sink); err != nil {
		return Sink{}, err
	}
	return sink, nil
}

// List returns every sink.
func (r *Registry) List() ([]Sink, error) {
	ids, err := r.fs.List("sinks")
	if err != nil {
		return nil, err
	}
	out := make([]Sink, 0, len(ids))
	for _, id := range ids {
		sink, err := r.Get(id)
		if err != nil {
			return nil, err
		}
		out = append(out, sink)
	}
	return out, nil
}

// SetFailures arms a sink to fail the next n deliveries.
func (r *Registry) SetFailures(id string, count int) error {
	sink, err := r.Get(id)
	if err != nil {
		return err
	}
	sink.FailuresRemaining = count
	sink.UpdatedAt = time.Now().UTC()
	return r.fs.WriteJSON(r.Path(id), sink)
}

// Deliver forwards lines to the sink, failing while armed.
func (r *Registry) Deliver(id string, lines []string) error {
	sink, err := r.Get(id)
	if err != nil {
		return err
	}
	if sink.FailuresRemaining > 0 {
		sink.FailuresRemaining--
		sink.UpdatedAt = time.Now().UTC()
		_ = r.fs.WriteJSON(r.Path(id), sink)
		return errors.New("receiver returned 5xx")
	}
	sink.Delivered += len(lines)
	sink.UpdatedAt = time.Now().UTC()
	return r.fs.WriteJSON(r.Path(id), sink)
}
