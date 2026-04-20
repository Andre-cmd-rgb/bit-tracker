package diary

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/andre-cmd-rgb/bit-tracker/internal/db"
)

func newRepo(t *testing.T) *Repo {
	t.Helper()
	store, err := db.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	return NewRepo(store)
}

func TestGetOrCreateTodayAutoCreates(t *testing.T) {
	r := newRepo(t)
	now := time.Date(2026, 4, 20, 7, 14, 0, 0, time.UTC)
	e, created, err := r.GetOrCreateToday(now)
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Fatal("expected created=true on first call")
	}
	if !strings.Contains(e.RawText, "monday, 20 april 2026") {
		t.Errorf("intro missing: %q", e.RawText)
	}
	if !strings.Contains(e.RawText, "started at 07:14") {
		t.Errorf("timestamp missing")
	}

	_, created2, err := r.GetOrCreateToday(now)
	if err != nil {
		t.Fatal(err)
	}
	if created2 {
		t.Fatal("second call should not create")
	}
}

func TestSaveScoresAndPersists(t *testing.T) {
	r := newRepo(t)
	now := time.Date(2026, 4, 20, 7, 14, 0, 0, time.UTC)
	e, _, err := r.GetOrCreateToday(now)
	if err != nil {
		t.Fatal(err)
	}
	e.Mood = 9
	e.StudyMinutes = 180
	e.ScrollMinutes = 5
	e.ProjectMinutes = 120
	e.Tags = []string{"focus", "coding"}
	if err := r.Save(&e); err != nil {
		t.Fatal(err)
	}

	back, err := r.GetByDate(now)
	if err != nil {
		t.Fatal(err)
	}
	if !back.IsGoodDay || back.IsBadDay {
		t.Fatalf("scoring didn't persist: good=%v bad=%v", back.IsGoodDay, back.IsBadDay)
	}
	if len(back.Tags) != 2 || back.Tags[0] != "focus" {
		t.Fatalf("tags not persisted: %v", back.Tags)
	}
}

func TestSearchByKeywordAndTag(t *testing.T) {
	r := newRepo(t)
	for i := 0; i < 3; i++ {
		d := time.Date(2026, 4, 18+i, 7, 0, 0, 0, time.UTC)
		e, _, _ := r.GetOrCreateToday(d)
		e.RawText += "\nshipped the thing"
		if i == 1 {
			e.Tags = []string{"focus"}
		}
		_ = r.Save(&e)
	}
	all, err := r.Search("shipped", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 {
		t.Fatalf("keyword search: got %d, want 3", len(all))
	}
	tagged, err := r.Search("", "focus")
	if err != nil {
		t.Fatal(err)
	}
	if len(tagged) != 1 {
		t.Fatalf("tag search: got %d, want 1", len(tagged))
	}
}
