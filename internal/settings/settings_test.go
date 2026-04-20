package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenCreatesDefaultFile(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "settings.json")); err != nil {
		t.Fatalf("settings.json missing: %v", err)
	}
	got := s.Get()
	if got.PetName != "bit" {
		t.Fatalf("default pet name %q want bit", got.PetName)
	}
	if got.PetCharacter != "penguin" {
		t.Fatalf("default character wrong")
	}
}

func TestUpdatePersists(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Update(func(s *Settings) {
		s.PetName = "pip"
		s.ThemeName = "green"
	}); err != nil {
		t.Fatal(err)
	}
	// Reload and verify.
	s2, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	got := s2.Get()
	if got.PetName != "pip" || got.ThemeName != "green" {
		t.Fatalf("not persisted: %+v", got)
	}
}

func TestMergesPartialConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(path, []byte(`{"pet_name":"zoe"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	got := s.Get()
	if got.PetName != "zoe" {
		t.Fatalf("partial name lost: %+v", got)
	}
	if got.PetCharacter != "penguin" {
		t.Fatalf("default character not merged: %+v", got)
	}
}
