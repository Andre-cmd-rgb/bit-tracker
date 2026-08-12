package diary

import "time"

// AISource wraps a Repo to present the read-only shape the AI layer expects.
// Its methods match ai.DataSource structurally, so the ai package can consume
// it without diary needing to import ai (which would be a cycle).
type AISource struct{ repo *Repo }

// NewAISource returns an adapter over the given repo.
func NewAISource(r *Repo) *AISource { return &AISource{repo: r} }

// All returns every entry, oldest first.
func (s *AISource) All() ([]Entry, error) {
	return s.repo.List(time.Time{}, time.Time{})
}

// Recent returns up to n most recent entries, oldest first. n <= 0 means all.
func (s *AISource) Recent(n int) ([]Entry, error) {
	all, err := s.repo.List(time.Time{}, time.Time{})
	if err != nil {
		return nil, err
	}
	if n <= 0 || len(all) <= n {
		return all, nil
	}
	return all[len(all)-n:], nil
}

// Range returns entries whose date falls in [from, to]. Zero times are open.
func (s *AISource) Range(from, to time.Time) ([]Entry, error) {
	return s.repo.List(from, to)
}

// ByDate returns the entry for that day or ErrNotFound.
func (s *AISource) ByDate(t time.Time) (Entry, error) {
	return s.repo.GetByDate(t)
}

// Search matches keyword across body text / project fields, with optional tag.
func (s *AISource) Search(keyword, tag string) ([]Entry, error) {
	return s.repo.Search(keyword, tag)
}
