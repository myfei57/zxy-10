package audit

// List returns the most recent audit entries, newest first.
func (s *Service) List(limit int) ([]Entry, error) {
	ids, err := s.fs.Latest("audit", limit)
	if err != nil {
		return nil, err
	}
	out := make([]Entry, 0, len(ids))
	for _, id := range ids {
		var entry Entry
		if err := s.fs.ReadJSON(s.fs.Path("audit", id), &entry); err != nil {
			return nil, err
		}
		out = append(out, entry)
	}
	return out, nil
}
