package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/copyagent/copyagentd/internal/config"
)

const (
	DefaultLogMaxSize = 10 * 1024 * 1024
	LogFileEnv        = "COPYAGENT_LOG_FILE"
	LogMaxSizeEnv     = "COPYAGENT_LOG_MAX_SIZE"
	ConfigPathEnv     = "COPYAGENT_CONFIG"
)

type Config struct {
	BinaryPath string
	ConfigPath string
	WorkDir    string
	LogFile    string
	LogMaxSize int64
	EnvPATH    string
}

type ServiceStatus struct {
	Installed bool
	Running   bool
	PID       int
	Platform  string
}

type Meta struct {
	LogFile     string `json:"log_file"`
	LogMaxSize  int64  `json:"log_max_size"`
	WorkDir     string `json:"work_dir"`
	BinaryPath  string `json:"binary_path"`
	ConfigPath  string `json:"config_path"`
	InstalledAt string `json:"installed_at"`
}

func DefaultLogFile() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".copyagent", "logs", "copyagentd.log")
}

func DefaultDataDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".copyagent")
}

func ConfigPath() string {
	if path := os.Getenv(ConfigPathEnv); path != "" {
		return path
	}
	return config.DefaultPath()
}

func Resolve(cfg *Config) error {
	if cfg.BinaryPath == "" {
		exe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("cannot detect binary path: %w", err)
		}
		if real, err := filepath.EvalSymlinks(exe); err == nil {
			exe = real
		}
		cfg.BinaryPath = exe
	}
	if cfg.ConfigPath == "" {
		cfg.ConfigPath = ConfigPath()
	}
	if cfg.WorkDir == "" {
		if cfg.ConfigPath != "" {
			cfg.WorkDir = filepath.Dir(cfg.ConfigPath)
		} else {
			wd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("cannot detect working directory: %w", err)
			}
			cfg.WorkDir = wd
		}
	}
	if cfg.LogFile == "" {
		cfg.LogFile = DefaultLogFile()
	}
	if cfg.LogMaxSize <= 0 {
		cfg.LogMaxSize = DefaultLogMaxSize
	}
	if cfg.EnvPATH == "" {
		cfg.EnvPATH = os.Getenv("PATH")
	}
	if cfg.EnvPATH == "" {
		cfg.EnvPATH = defaultEnvPATH()
	}
	return nil
}

func defaultEnvPATH() string {
	parts := make([]string, 0, 9)
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		parts = append(parts,
			filepath.Join(home, ".local", "bin"),
			filepath.Join(home, "bin"),
		)
	}
	parts = append(parts,
		"/opt/homebrew/bin",
		"/opt/homebrew/sbin",
		"/usr/local/bin",
		"/usr/bin",
		"/bin",
		"/usr/sbin",
		"/sbin",
		"/Applications/Codex.app/Contents/Resources",
	)

	seen := make(map[string]struct{}, len(parts))
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		filtered = append(filtered, part)
	}
	return strings.Join(filtered, ":")
}

func metaPath() string {
	return filepath.Join(DefaultDataDir(), "daemon.json")
}

func SaveMeta(meta *Meta) error {
	if err := os.MkdirAll(filepath.Dir(metaPath()), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(metaPath(), data, 0o644)
}

func LoadMeta() (*Meta, error) {
	data, err := os.ReadFile(metaPath())
	if err != nil {
		return nil, err
	}
	var meta Meta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func RemoveMeta() {
	_ = os.Remove(metaPath())
}

func NowISO() string {
	return time.Now().Format(time.RFC3339)
}
