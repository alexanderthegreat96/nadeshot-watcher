package functions

import (
"os"
"path/filepath"
"sync"
"testing"
"time"

"github.com/alexanderthegreat96/nadeshot-watcher/config"
)

func TestFileExists_ExistingFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_file_*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	exists, err := FileExists(tmpFile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected file to exist")
	}
}

func TestFileExists_NonExistentFile(t *testing.T) {
	exists, err := FileExists("/nonexistent/path/to/file.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected file to not exist")
	}
}

func TestFileExists_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	exists, err := FileExists(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected directory to exist")
	}
}

func TestIsPythonInstalled_InvalidCommand(t *testing.T) {
	result := IsPythonInstalled("nonexistent_command_12345")
	if result {
		t.Error("expected false for nonexistent command")
	}
}

func TestNewAppRunner(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.ExeDir = "/test"

	runner := NewAppRunner(cfg)

	if runner == nil {
		t.Fatal("expected non-nil AppRunner")
	}
	if runner.cfg != cfg {
		t.Error("expected cfg to be set")
	}
	if runner.isRunning {
		t.Error("expected isRunning to be false initially")
	}
}

func TestAppRunner_Shutdown(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.ExeDir = t.TempDir()

	runner := NewAppRunner(cfg)

	runner.Shutdown()
	runner.Shutdown()
	runner.Shutdown()
}

func TestAppRunner_ShutdownOnce(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.ExeDir = t.TempDir()

	runner := NewAppRunner(cfg)

	shutdownCount := 0
	var mu sync.Mutex

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runner.Shutdown()
			mu.Lock()
			shutdownCount++
			mu.Unlock()
		}()
	}
	wg.Wait()

	if shutdownCount != 10 {
		t.Errorf("expected 10 calls to Shutdown, got %d", shutdownCount)
	}
}

func TestWatchRegular_InvalidDirectory(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.ExeDir = "/nonexistent/directory/path"

	runner := NewAppRunner(cfg)

	err := WatchRegular(cfg, runner)
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

func TestWatchRegular_ValidDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.ExeDir = tmpDir

	runner := NewAppRunner(cfg)

	err := WatchRegular(cfg, runner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
}

func TestWatchPolling_InvalidDirectory(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.ExeDir = "/nonexistent/directory/path"

	runner := NewAppRunner(cfg)

	err := WatchPolling(cfg, runner)
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

func TestWatchPolling_ValidDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.ExeDir = tmpDir

	runner := NewAppRunner(cfg)

	err := WatchPolling(cfg, runner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
}

func TestAppRunner_ConcurrentRunApp(t *testing.T) {
	tmpDir := t.TempDir()

	scriptPath := filepath.Join(tmpDir, "main.py")
	if err := os.WriteFile(scriptPath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create script: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.ExeDir = tmpDir
	cfg.PythonCommand = "true"
	cfg.PythonArgs = []string{}
	cfg.DebounceMs = 10

	runner := NewAppRunner(cfg)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runner.RunApp()
		}()
	}
	wg.Wait()
}

func TestFileExists_EmptyPath(t *testing.T) {
	exists, err := FileExists("")
	if err == nil && exists {
		t.Error("expected false or error for empty path")
	}
}

func TestShouldIgnore_PartialMatch(t *testing.T) {
	cfg := &config.Config{
		IgnorePatterns: []string{"cache"},
	}

	if !cfg.ShouldIgnore("/project/__pycache__/file.py") {
		t.Error("expected __pycache__ to match 'cache' pattern")
	}
}

func TestShouldWatch_NoExtension(t *testing.T) {
	cfg := &config.Config{
		WatchExtensions: []string{".py"},
	}

	if cfg.ShouldWatch("Makefile") {
		t.Error("expected false for file without extension")
	}
}
