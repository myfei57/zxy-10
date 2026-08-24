package store

import (
	"os"
	"path/filepath"
)

// Kinds enumerates the record directories the control plane persists. Keeping
// them explicit here makes the on-disk layout discoverable and lets Open
// pre-create every directory on startup.
var Kinds = []string{
	"sources",
	"pipelines",
	"rules",
	"checkpoints",
	"batches",
	"deadletters",
	"subscriptions",
	"sinks",
	"audit",
}

// Store is the aggregate file-backed persistence root used by the domain
// services. It owns the FileStore and guarantees the record directories
// exist before any service writes.
type Store struct {
	FS *FileStore
}

// Open creates the record directories under root and returns the aggregate
// Store. An existing data directory is reused as-is.
func Open(root string) (*Store, error) {
	for _, kind := range Kinds {
		if err := os.MkdirAll(filepath.Join(root, kind), 0o755); err != nil {
			return nil, err
		}
	}
	return &Store{FS: NewFileStore(root)}, nil
}
