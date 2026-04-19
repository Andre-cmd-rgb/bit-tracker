package config

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
)

type Config struct {
	ModelName    string `toml:"model_name"`
	ModelPath    string `toml:"model_path"`
	LlamaBin     string `toml:"llama_bin"`
	DBPath       string `toml:"db_path"`
	DataDir      string `toml:"data_dir"`
	ConfigDir    string `toml:"config_dir"`
	ConfigFile   string `toml:"-"`
	FirstRunDone bool   `toml:"first_run_done"`
	Username     string `toml:"username"`
}

func xdg(env, fallback string) string {
	if v := os.Getenv(env); v != "" {
		return v
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, fallback)
}

func Load() (*Config, error) {
	configDir := filepath.Join(xdg("XDG_CONFIG_HOME", ".config"), "bit-tracker")
	dataDir := filepath.Join(xdg("XDG_DATA_HOME", ".local/share"), "bit-tracker")

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(dataDir, "models"), 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(dataDir, "bin"), 0o755); err != nil {
		return nil, err
	}

	cfgPath := filepath.Join(configDir, "config.toml")
	cfg := &Config{
		ConfigDir:  configDir,
		ConfigFile: cfgPath,
		DataDir:    dataDir,
		DBPath:     filepath.Join(dataDir, "bit-tracker.db"),
		Username:   "you",
	}

	if b, err := os.ReadFile(cfgPath); err == nil {
		_ = toml.Unmarshal(b, cfg)
		cfg.ConfigDir = configDir
		cfg.ConfigFile = cfgPath
		cfg.DataDir = dataDir
		if cfg.DBPath == "" {
			cfg.DBPath = filepath.Join(dataDir, "bit-tracker.db")
		}
	}

	if cfg.LlamaBin == "" {
		bin := "llama-server"
		if runtime.GOOS == "windows" {
			bin += ".exe"
		}
		cfg.LlamaBin = filepath.Join(dataDir, "bin", bin)
	}

	return cfg, nil
}

func (c *Config) Save() error {
	f, err := os.Create(c.ConfigFile)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(c)
}
