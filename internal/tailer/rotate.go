package tailer

import "os"

// DetectRotation reports whether the file at path is smaller than the cursor
// offset, which means the log rotated and the cursor must rewind to the new
// file start.
func DetectRotation(path string, offset int64) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.Size() < offset, nil
}
