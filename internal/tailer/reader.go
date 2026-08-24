package tailer

import (
	"os"
	"strings"
)

// ReadResult is one chunk of lines read from a source file together with the
// new byte offset.
type ReadResult struct {
	Lines  []string
	Offset int64
}

// ReadLines reads lines from path starting at offset. It never rewinds: the
// returned offset is where the next read should resume.
func ReadLines(path string, offset int64) (ReadResult, error) {
	file, err := os.Open(path)
	if err != nil {
		return ReadResult{}, err
	}
	defer file.Close()
	if offset < 0 {
		offset = 0
	}
	if _, err := file.Seek(offset, 0); err != nil {
		return ReadResult{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ReadResult{}, err
	}
	text := string(data)
	if int64(len(text)) < offset {
		return ReadResult{}, os.ErrInvalid
	}
	chunk := text[offset:]
	lines := strings.Split(chunk, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return ReadResult{Lines: lines, Offset: int64(len(text))}, nil
}
