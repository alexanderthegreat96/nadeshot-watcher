package functions

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"time"

	"github.com/alexanderthegreat96/nadeshot-watcher/config"
	"github.com/fsnotify/fsnotify"
	"github.com/radovskyb/watcher"
)

type AppRunner struct {
	cfg *config.Config

	mu           sync.Mutex
	ctx          context.Context
	cancel       context.CancelFunc
	isRunning    bool
	lastRestart  time.Time
	shutdownOnce sync.Once
	doneCh       chan struct{} // signals when process has exited
}

func NewAppRunner(cfg *config.Config) *AppRunner {
	return &AppRunner{
		cfg: cfg,
	}
}

func (r *AppRunner) RunApp() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if time.Since(r.lastRestart) < time.Duration(r.cfg.DebounceMs)*time.Millisecond {
		return
	}

	if r.isRunning && r.cancel != nil {
		log.Println("Restarting the previous app instance...")
		r.cancel()

		// Wait for the old process to actually exit
		if r.doneCh != nil {
			r.mu.Unlock()
			<-r.doneCh
			r.mu.Lock()
		}
	}

	r.ctx, r.cancel = context.WithCancel(context.Background())
	r.doneCh = make(chan struct{})
	r.lastRestart = time.Now()
	r.isRunning = true

	go r.runProcess(r.ctx, r.doneCh)
}

func (r *AppRunner) runProcess(ctx context.Context, doneCh chan struct{}) {
	defer close(doneCh)

	mainPath := r.cfg.BootFilePath()

	args := make([]string, 0, len(r.cfg.PythonArgs)+1+len(r.cfg.ScriptArgs))
	args = append(args, r.cfg.PythonArgs...)
	args = append(args, mainPath)
	args = append(args, r.cfg.ScriptArgs...)

	cmd := exec.CommandContext(ctx, r.cfg.PythonCommand, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = r.cfg.ExeDir

	// Go 1.20+ context cancelation requires you to specify it
	cmd.Cancel = func() error {
		log.Println("Killing process...")
		return cmd.Process.Kill()
	}

	err := cmd.Run()
	if err != nil {
		if ctx.Err() == nil {
			log.Printf("App exited with error: %v", err)
		} else {
			log.Println("Previous instance stopped.")
		}
	}

	r.mu.Lock()
	r.isRunning = false
	r.mu.Unlock()
}

func (r *AppRunner) Shutdown() {
	r.shutdownOnce.Do(func() {
		r.mu.Lock()
		defer r.mu.Unlock()

		log.Println("Shutting down application...")
		if r.cancel != nil {
			r.cancel()
		}
	})
}

func (r *AppRunner) SetupSignalHandler() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %v", sig)
		r.Shutdown()
		os.Exit(0)
	}()
}

func WatchRegular(cfg *config.Config, runner *AppRunner) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

	if err := w.Add(cfg.ExeDir); err != nil {
		return fmt.Errorf("failed to watch directory: %w", err)
	}

	err = filepath.Walk(cfg.ExeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("Error walking path:", err)
			return nil
		}
		if info.IsDir() {
			if !cfg.ShouldIgnore(path) {
				if err := w.Add(path); err != nil {
					log.Println("Error adding directory to watcher:", err)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Println("Error walking path:", err)
	}

	go func() {
		for {
			select {
			case event, ok := <-w.Events:
				if !ok {
					return
				}
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) != 0 {
					if !cfg.ShouldIgnore(event.Name) && cfg.ShouldWatch(event.Name) {
						log.Printf("Change detected: %s", event.Name)
						runner.RunApp()
					}
				}
			case err, ok := <-w.Errors:
				if !ok {
					return
				}
				log.Println("Watcher error:", err)
			}
		}
	}()

	return nil
}

func WatchPolling(cfg *config.Config, runner *AppRunner) error {
	w := watcher.New()
	w.SetMaxEvents(1)
	w.FilterOps(watcher.Write, watcher.Create, watcher.Remove)

	if err := w.AddRecursive(cfg.ExeDir); err != nil {
		return fmt.Errorf("failed to add directory to watcher: %w", err)
	}

	for path := range w.WatchedFiles() {
		if cfg.ShouldIgnore(path) {
			w.Remove(path)
		}
	}

	go func() {
		for {
			select {
			case event := <-w.Event:
				if !cfg.ShouldIgnore(event.Path) && cfg.ShouldWatch(event.Path) {
					log.Printf("Change detected: %s", event.Path)
					runner.RunApp()
				}
			case err := <-w.Error:
				log.Println("Watcher error:", err)
			case <-w.Closed:
				return
			}
		}
	}()

	go func() {
		if err := w.Start(time.Second); err != nil {
			log.Printf("Watcher start error: %v", err)
		}
	}()

	return nil
}

func FileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func IsPythonInstalled(pythonCmd string) bool {
	cmd := exec.Command(pythonCmd, "--version")
	err := cmd.Run()
	return err == nil
}
