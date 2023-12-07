package functions

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/radovskyb/watcher"
)

var (
	mu           sync.Mutex
	isBotRunning bool
	ctx          context.Context
	cancel       context.CancelFunc
	stopCh       chan struct{}
)

// uses watcher and doesn't rely on FS events
// for docker usage obviosly
func WatcherWatchPath(w *watcher.Watcher, path string) {
	// Add the current path.
	if err := w.AddRecursive(path); err != nil {
		log.Fatal(err)
		return
	}

	// Start the watcher
	go func() {
		for {
			select {
			case event := <-w.Event:
				if event.Op&watcher.Write == watcher.Write {
					if !containsIgnorePath(event.Path, "__pycache__") {
						log.Println("Event:", event)
						go RunBot("main.py")
					}
				}
			case err := <-w.Error:
				log.Println("Error:", err)
			case <-w.Closed:
				return
			}
		}
	}()

	// Start the watcher
	if err := w.Start(1); err != nil {
		fmt.Println("Error starting watcher:", err)
		fmt.Println("Press ENTER to exit...")
		fmt.Scanln()
		os.Exit(1)
	}

}
func RunBot(scriptPath string) {
	mu.Lock()
	defer mu.Unlock()

	if isBotRunning {
		log.Println("Restarting the previous bot instance.")
		cancel()
	}

	ctx, cancel = context.WithCancel(context.Background())

	go func() {
		cmd := exec.CommandContext(ctx, "python", scriptPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			//log.Printf("Error running bot: %v", err)
			log.Printf("Previous instance stopped...")
		}

		// The bot has finished running.
		mu.Lock()
		defer mu.Unlock()
		isBotRunning = true
	}()

	isBotRunning = true
}

func WatchPath(watcher *fsnotify.Watcher, path string) {
	// Add the current path.
	err := watcher.Add(path)
	if err != nil {
		log.Fatal(err)
		fmt.Println("Press ENTER to exit...")
		fmt.Scanln()
		return
	}

	err = filepath.Walk(path, func(subpath string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("Error walking path:", err)
			return nil
		}
		if info.IsDir() {
			if !containsIgnorePath(subpath, "__pycache__") {
				err := watcher.Add(subpath)
				if err != nil {
					log.Println("Error adding subdirectory to watcher:", err)
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
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
					log.Println("event", event.Name)

					// Run the bot in a separate goroutine.
					go RunBot("main.py")
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Error:", err)
			}
		}
	}()
}

func containsIgnorePath(path, ignorePath string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		fmt.Println("Error getting absolute path:", err)
		return false
	}

	absIgnorePath, err := filepath.Abs(ignorePath)
	if err != nil {
		fmt.Println("Error getting absolute ignore path:", err)
		return false
	}

	return strings.HasPrefix(absPath, absIgnorePath)
}
