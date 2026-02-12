package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexanderthegreat96/envparser/v3"
)

type WatchMode string

const (
	ModeRegular WatchMode = "regular"
	ModePolling WatchMode = "docker"
)

type Config struct {
	Mode            WatchMode
	BootFile        string
	PythonCommand   string
	PythonArgs      []string
	ScriptArgs      []string
	DebounceMs      int
	IgnorePatterns  []string
	WatchExtensions []string
	ExeDir          string
}

func DefaultConfig() *Config {
	return &Config{
		Mode:            ModeRegular,
		BootFile:        "main.py",
		PythonCommand:   "python",
		PythonArgs:      []string{"-B"}, // -B = don't write .pyc files
		ScriptArgs:      []string{},
		DebounceMs:      500,
		IgnorePatterns:  []string{"__pycache__", ".git", ".venv", "venv", "node_modules", ".idea", ".vscode"},
		WatchExtensions: []string{".py"},
	}
}

func LoadConfig(exeDir string) (*Config, error) {
	cfg := DefaultConfig()
	cfg.ExeDir = exeDir

	configPath := filepath.Join(exeDir, "watcher.ini")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if cwd, cwdErr := os.Getwd(); cwdErr == nil {
			cwdConfigPath := filepath.Join(cwd, "watcher.ini")
			if _, statErr := os.Stat(cwdConfigPath); statErr == nil {
				configPath = cwdConfigPath
				cfg.ExeDir = cwd
			}
		}
	}

	env := envparser.NewEnvParser(configPath, false)
	if env.EnvError != nil {
		if strings.Contains(env.GetError(), "no such file") ||
			strings.Contains(env.GetError(), "does not exist") {
			return cfg, nil
		}
		return nil, fmt.Errorf("error loading config: %s", env.GetError())
	}

	if mode, err := env.GetValue("MODE", "string", string(cfg.Mode)); err == nil {
		modeStr := strings.ToLower(fmt.Sprintf("%v", mode))
		switch modeStr {
		case "regular", "fsnotify", "native":
			cfg.Mode = ModeRegular
		case "docker", "polling", "poll":
			cfg.Mode = ModePolling
		}
	}

	if bootFile, err := env.GetValue("BOOT_FILE", "string", cfg.BootFile); err == nil {
		bf := fmt.Sprintf("%v", bootFile)
		if strings.HasSuffix(bf, ".py") {
			cfg.BootFile = bf
		}
	}

	if python, err := env.GetValue("PYTHON", "string", cfg.PythonCommand); err == nil {
		cfg.PythonCommand = fmt.Sprintf("%v", python)
	}

	if pythonArgs, err := env.GetValue("PYTHON_ARGS", "list", nil); err == nil {
		if argsList, ok := pythonArgs.([]any); ok && len(argsList) > 0 {
			cfg.PythonArgs = make([]string, 0, len(argsList))
			for _, item := range argsList {
				arg := strings.TrimSpace(fmt.Sprintf("%v", item))
				if arg != "" {
					cfg.PythonArgs = append(cfg.PythonArgs, arg)
				}
			}
		}
	}

	if scriptArgs, err := env.GetValue("SCRIPT_ARGS", "list", nil); err == nil {
		if argsList, ok := scriptArgs.([]any); ok && len(argsList) > 0 {
			cfg.ScriptArgs = make([]string, 0, len(argsList))
			for _, item := range argsList {
				arg := strings.TrimSpace(fmt.Sprintf("%v", item))
				if arg != "" {
					cfg.ScriptArgs = append(cfg.ScriptArgs, arg)
				}
			}
		}
	}

	if debounce, err := env.GetValue("DEBOUNCE_MS", "int", cfg.DebounceMs); err == nil {
		if ms, ok := debounce.(int); ok && ms >= 0 {
			cfg.DebounceMs = ms
		}
	}

	if ignore, err := env.GetValue("IGNORE", "list", nil); err == nil {
		if ignoreList, ok := ignore.([]any); ok && len(ignoreList) > 0 {
			cfg.IgnorePatterns = make([]string, 0, len(ignoreList))
			for _, item := range ignoreList {
				pattern := strings.TrimSpace(fmt.Sprintf("%v", item))
				if pattern != "" {
					cfg.IgnorePatterns = append(cfg.IgnorePatterns, pattern)
				}
			}
		}
	}

	if watch, err := env.GetValue("WATCH_EXTENSIONS", "list", nil); err == nil {
		if watchList, ok := watch.([]any); ok && len(watchList) > 0 {
			cfg.WatchExtensions = make([]string, 0, len(watchList))
			for _, item := range watchList {
				ext := strings.TrimSpace(fmt.Sprintf("%v", item))
				if ext != "" {
					if !strings.HasPrefix(ext, ".") {
						ext = "." + ext
					}
					cfg.WatchExtensions = append(cfg.WatchExtensions, ext)
				}
			}
		}
	}

	return cfg, nil
}

func (c *Config) ShouldIgnore(path string) bool {
	for _, pattern := range c.IgnorePatterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

func (c *Config) ShouldWatch(path string) bool {
	if len(c.WatchExtensions) == 0 {
		return true
	}

	ext := strings.ToLower(filepath.Ext(path))
	for _, watchExt := range c.WatchExtensions {
		if ext == strings.ToLower(watchExt) {
			return true
		}
	}
	return false
}

func (c *Config) BootFilePath() string {
	return filepath.Join(c.ExeDir, c.BootFile)
}

func (c *Config) String() string {
	return fmt.Sprintf(
		"Configuration Status:\n"+
			"  %-16s %s\n"+
			"  %-16s %s\n"+
			"  %-16s %s\n"+
			"  %-16s %v\n"+
			"  %-16s %v\n"+
			"  %-16s %dms\n"+
			"  %-16s %v\n"+
			"  %-16s %v",
		"Mode:", c.Mode,
		"Boot File:", c.BootFile,
		"Python:", c.PythonCommand,
		"Python Args:", c.PythonArgs,
		"Script Args:", c.ScriptArgs,
		"Debounce:", c.DebounceMs,
		"Ignore Patterns:", c.IgnorePatterns,
		"Watch Exts:", c.WatchExtensions,
	)
}
