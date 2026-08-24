package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ErrNotFound is returned when a requested record does not exist on disk.
var ErrNotFound = errors.New("store: record not found")

// FileStore persists JSON records as individual files under a data directory.
// Each record kind maps to a subdirectory and each record id maps to
// <kind>/<id>.json. Writes go through a temporary file first and are renamed
// into place, so a completed write is atomic and a crashed write never leaves
// a partially readable record.
type FileStore struct {
	root string
}

// NewFileStore creates a FileStore rooted at the given directory.
func NewFileStore(root string) *FileStore {
	return &FileStore{root: root}
}

// Path resolves the JSON file for a record kind and id.
func (f *FileStore) Path(kind, id string) string {
	return filepath.Join(f.root, kind, id+".json")
}

// KindDir resolves the directory holding a record kind.
func (f *FileStore) KindDir(kind string) string {
	return filepath.Join(f.root, kind)
}

// WriteJSON serializes value and durably writes it to path.
func (f *FileStore) WriteJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return f.WriteBytes(path, data)
}

// WriteBytes writes raw bytes durably to path through a temp file and rename.
func (f *FileStore) WriteBytes(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	_, writeErr := tmp.Write(data)
	closeErr := tmp.Close()
	if writeErr != nil || closeErr != nil {
		_ = os.Remove(tmpName)
		if writeErr != nil {
			return writeErr
		}
		return closeErr
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return nil
}

// ReadJSON reads and decodes the record at path. Missing files map to
// ErrNotFound so callers can distinguish absent state from corrupt state.
func (f *FileStore) ReadJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return err
	}
	if err := json.Unmarshal(data, value); err != nil {
		return fmt.Errorf("store: decode %s: %w", path, err)
	}
	return nil
}

// List returns the ids of all JSON records in a kind, sorted lexically.
func (f *FileStore) List(kind string) ([]string, error) {
	entries, err := os.ReadDir(f.KindDir(kind))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".json") {
			ids = append(ids, strings.TrimSuffix(name, ".json"))
		}
	}
	sort.Strings(ids)
	return ids, nil
}

// Latest returns record ids in a kind ordered by modification time, newest
// first, limited to at most limit entries.
func (f *FileStore) Latest(kind string, limit int) ([]string, error) {
	ids, err := f.List(kind)
	if err != nil {
		return nil, err
	}
	type entry struct {
		id    string
		mtime time.Time
	}
	entries := make([]entry, 0, len(ids))
	for _, id := range ids {
		info, err := os.Stat(f.Path(kind, id))
		if err != nil {
			continue
		}
		entries = append(entries, entry{id: id, mtime: info.ModTime()})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].mtime.After(entries[j].mtime)
	})
	if len(entries) > limit {
		entries = entries[:limit]
	}
	out := make([]string, 0, len(entries))
	for _, item := range entries {
		out = append(out, item.id)
	}
	return out, nil
}
