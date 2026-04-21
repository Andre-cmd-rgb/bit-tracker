// Package settings holds user-configurable preferences: pet customisation,
// colour theme, and the active local model.
package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Settings is the full user-editable config.
type Settings struct {
	// Pet
	PetName      string `json:"pet_name"`
	PetCharacter string `json:"pet_character"` // penguin | robot | cat | bunny | ghost | fox | dragon | owl
	PetHat       string `json:"pet_hat"`       // none | crown | tophat | propeller | halo | wizard | beanie | tinyduck | flower | antenna
	PetEyes      string `json:"pet_eyes"`      // dot | star | cross | circle | at | degree | minus | cute | sleepy | wink
	PetShiny     bool   `json:"pet_shiny"`

	// Theme
	ThemeName string `json:"theme_name"` // purple | green | amber | cyan | rose

	// AI
	ActiveModel string `json:"active_model"` // id of chosen gguf (resolves to catalog filename)

	// Onboarding
	SetupComplete bool `json:"setup_complete"`
}

// Default returns sane defaults when no settings file exists.
func Default() Settings {
	return Settings{
		PetName:      "bit",
		PetCharacter: "penguin",
		PetHat:       "crown",
		PetEyes:      "cross",
		PetShiny:     true,
		ThemeName:    "purple",
		ActiveModel:  "",
	}
}

type Store struct {
	mu    sync.Mutex
	path  string
	cur   Settings
	fresh bool // true when settings.json did not exist at Open time
}

// Open loads settings.json from dir, writing defaults if missing.
func Open(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	p := filepath.Join(dir, "settings.json")
	s := &Store{path: p, cur: Default()}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			s.fresh = true
			if err := s.save(); err != nil {
				return nil, err
			}
			return s, nil
		}
		return nil, err
	}
	var parsed Settings
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("settings: %w", err)
	}
	// Merge with defaults so new fields are populated.
	merged := Default()
	if parsed.PetName != "" {
		merged.PetName = parsed.PetName
	}
	if parsed.PetCharacter != "" {
		merged.PetCharacter = parsed.PetCharacter
	}
	if parsed.PetHat != "" {
		merged.PetHat = parsed.PetHat
	}
	if parsed.PetEyes != "" {
		merged.PetEyes = parsed.PetEyes
	}
	merged.PetShiny = parsed.PetShiny
	if parsed.ThemeName != "" {
		merged.ThemeName = parsed.ThemeName
	}
	merged.ActiveModel = parsed.ActiveModel
	merged.SetupComplete = parsed.SetupComplete
	s.cur = merged
	return s, nil
}

// IsFresh reports whether the store was created this launch (no prior
// settings.json on disk). Used to decide whether to show the setup flow.
func (s *Store) IsFresh() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.fresh
}

// Get returns a copy of current settings.
func (s *Store) Get() Settings {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cur
}

// Update mutates settings via fn and persists them.
func (s *Store) Update(fn func(*Settings)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(&s.cur)
	return s.save()
}

// Set replaces the whole settings struct and persists it.
func (s *Store) Set(next Settings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cur = next
	return s.save()
}

func (s *Store) Path() string { return s.path }

func (s *Store) save() error {
	data, err := json.MarshalIndent(s.cur, "", "  ")
	if err != nil {
		return err
	}
	// Atomic write.
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
