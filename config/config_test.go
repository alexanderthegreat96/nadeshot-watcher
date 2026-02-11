package config

import (
"os"
"path/filepath"
"reflect"
"strings"
"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Mode != ModeRegular {
		t.Errorf("expected Mode=%s, got %s", ModeRegular, cfg.Mode)
	}
	if cfg.BootFile != "main.py" {
		t.Errorf("expected BootFile=main.py, got %s", cfg.BootFile)
	}
	if cfg.PythonCommand != "python" {
		t.Errorf("expected PythonCommand=python, got %s", cfg.PythonCommand)
	}
	if cfg.DebounceMs != 500 {
		t.Errorf("expected DebounceMs=500, got %d", cfg.DebounceMs)
	}
	if len(cfg.PythonArgs) != 1 || cfg.PythonArgs[0] != "-B" {
		t.Errorf("expected PythonArgs=[-B], got %v", cfg.PythonArgs)
	}
	if len(cfg.ScriptArgs) != 0 {
		t.Errorf("expected ScriptArgs=[], got %v", cfg.ScriptArgs)
	}
	if len(cfg.IgnorePatterns) == 0 {
		t.Error("expected IgnorePatterns to have default values")
	}
	if len(cfg.WatchExtensions) != 1 || cfg.WatchExtensions[0] != ".py" {
		t.Errorf("expected WatchExtensions=[.py], got %v", cfg.WatchExtensions)
	}
}

func TestLoadConfig_NoConfigFile(t *testing.T) {
	tmpDir := t.TempDir()

	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Mode != ModeRegular {
		t.Errorf("expected default Mode=%s, got %s", ModeRegular, cfg.Mode)
	}
	if cfg.BootFile != "main.py" {
		t.Errorf("expected default BootFile=main.py, got %s", cfg.BootFile)
	}
}

func TestLoadConfig_WithConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "watcher.ini")

	configContent := "MODE=docker\nBOOT_FILE=app.py\nPYTHON=/usr/bin/python3\nPYTHON_ARGS=[-B, -u]\nSCRIPT_ARGS=[--debug, --port, 8080]\nDEBOUNCE_MS=200\nIGNORE=[__pycache__, .git, custom_ignore]\nWATCH_EXTENSIONS=[.py, .json]"

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Mode != ModePolling {
		t.Errorf("expected Mode=%s, got %s", ModePolling, cfg.Mode)
	}
	if cfg.BootFile != "app.py" {
		t.Errorf("expected BootFile=app.py, got %s", cfg.BootFile)
	}
	if cfg.PythonCommand != "/usr/bin/python3" {
		t.Errorf("expected PythonCommand=/usr/bin/python3, got %s", cfg.PythonCommand)
	}
	if cfg.DebounceMs != 200 {
		t.Errorf("expected DebounceMs=200, got %d", cfg.DebounceMs)
	}

	expectedPythonArgs := []string{"-B", "-u"}
	if !reflect.DeepEqual(cfg.PythonArgs, expectedPythonArgs) {
		t.Errorf("expected PythonArgs=%v, got %v", expectedPythonArgs, cfg.PythonArgs)
	}

	expectedScriptArgs := []string{"--debug", "--port", "8080"}
	if !reflect.DeepEqual(cfg.ScriptArgs, expectedScriptArgs) {
		t.Errorf("expected ScriptArgs=%v, got %v", expectedScriptArgs, cfg.ScriptArgs)
	}
}

func TestLoadConfig_ModeVariants(t *testing.T) {
	tests := []struct {
		modeValue    string
		expectedMode WatchMode
	}{
		{"regular", ModeRegular},
		{"REGULAR", ModeRegular},
		{"fsnotify", ModeRegular},
		{"native", ModeRegular},
		{"docker", ModePolling},
		{"DOCKER", ModePolling},
		{"polling", ModePolling},
		{"poll", ModePolling},
	}

	for _, tc := range tests {
		t.Run(tc.modeValue, func(t *testing.T) {
tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "watcher.ini")
			configContent := "MODE=" + tc.modeValue

			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				t.Fatalf("failed to create config file: %v", err)
			}

			cfg, err := LoadConfig(tmpDir)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cfg.Mode != tc.expectedMode {
				t.Errorf("for MODE=%s: expected %s, got %s", tc.modeValue, tc.expectedMode, cfg.Mode)
			}
		})
	}
}

func TestLoadConfig_BootFileValidation(t *testing.T) {
	tests := []struct {
		bootFile     string
		expectedBoot string
	}{
		{"app.py", "app.py"},
		{"my_script.py", "my_script.py"},
		{"notpython.txt", "main.py"},
	}

	for _, tc := range tests {
		t.Run(tc.bootFile, func(t *testing.T) {
tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "watcher.ini")
			configContent := "BOOT_FILE=" + tc.bootFile

			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				t.Fatalf("failed to create config file: %v", err)
			}

			cfg, err := LoadConfig(tmpDir)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cfg.BootFile != tc.expectedBoot {
				t.Errorf("for BOOT_FILE=%s: expected %s, got %s", tc.bootFile, tc.expectedBoot, cfg.BootFile)
			}
		})
	}
}

func TestShouldIgnore(t *testing.T) {
	cfg := &Config{
		IgnorePatterns: []string{"__pycache__", ".git", "node_modules"},
	}

	tests := []struct {
		path     string
		expected bool
	}{
		{"/project/__pycache__/file.pyc", true},
		{"/project/.git/config", true},
		{"/project/node_modules/package/index.js", true},
		{"/project/src/main.py", false},
		{"/project/app.py", false},
		{"__pycache__", true},
		{".git", true},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
result := cfg.ShouldIgnore(tc.path)
if result != tc.expected {
t.Errorf("ShouldIgnore(%s): expected %v, got %v", tc.path, tc.expected, result)
}
})
	}
}

func TestShouldIgnore_EmptyPatterns(t *testing.T) {
	cfg := &Config{
		IgnorePatterns: []string{},
	}

	if cfg.ShouldIgnore("/any/path/file.py") {
		t.Error("expected false when no ignore patterns")
	}
}

func TestShouldWatch(t *testing.T) {
	cfg := &Config{
		WatchExtensions: []string{".py", ".json"},
	}

	tests := []struct {
		path     string
		expected bool
	}{
		{"/project/main.py", true},
		{"/project/config.json", true},
		{"/project/script.PY", true},
		{"/project/data.JSON", true},
		{"/project/readme.md", false},
		{"/project/style.css", false},
		{"/project/image.png", false},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
result := cfg.ShouldWatch(tc.path)
if result != tc.expected {
t.Errorf("ShouldWatch(%s): expected %v, got %v", tc.path, tc.expected, result)
}
})
	}
}

func TestShouldWatch_EmptyExtensions(t *testing.T) {
	cfg := &Config{
		WatchExtensions: []string{},
	}

	if !cfg.ShouldWatch("/any/file.txt") {
		t.Error("expected true when no watch extensions (watch all)")
	}
	if !cfg.ShouldWatch("/any/file.py") {
		t.Error("expected true when no watch extensions (watch all)")
	}
}

func TestBootFilePath(t *testing.T) {
	cfg := &Config{
		ExeDir:   "/home/user/project",
		BootFile: "main.py",
	}

	expected := filepath.Join("/home/user/project", "main.py")
	result := cfg.BootFilePath()

	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestString(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ExeDir = "/test"

	str := cfg.String()

	if !strings.Contains(str, "Mode:") {
		t.Error("String() should contain Mode")
	}
	if !strings.Contains(str, "Boot File:") {
		t.Error("String() should contain Boot File")
	}
	if !strings.Contains(str, "Python:") {
		t.Error("String() should contain Python")
	}
	if !strings.Contains(str, "Debounce:") {
		t.Error("String() should contain Debounce")
	}
}

func TestWatchModeConstants(t *testing.T) {
	if ModeRegular != "regular" {
		t.Errorf("ModeRegular should be 'regular', got %s", ModeRegular)
	}
	if ModePolling != "docker" {
		t.Errorf("ModePolling should be 'docker', got %s", ModePolling)
	}
}
