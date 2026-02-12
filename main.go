package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alexanderthegreat96/nadeshot-watcher/config"
	"github.com/alexanderthegreat96/nadeshot-watcher/functions"
	"github.com/common-nighthawk/go-figure"
)

func main() {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("Error getting executable path: %s\n", err)
		return
	}
	exeDir := filepath.Dir(exePath)

	myFigure := figure.NewColorFigure("Nadeshot Watcher", "", "green", true)
	myFigure.Print()
	fmt.Println()

	cfg, err := config.LoadConfig(exeDir)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		waitForExit()
		return
	}

	fmt.Printf("Config: %s\n\n", cfg)

	if !functions.IsPythonInstalled(cfg.PythonCommand) {
		fmt.Printf("Python interpreter '%s' not found.\n", cfg.PythonCommand)
		fmt.Println("Please install Python or update the PYTHON setting in watcher.ini")
		waitForExit()
		return
	}

	found, err := functions.FileExists(cfg.BootFilePath())
	if err != nil {
		fmt.Printf("Error checking boot file: %v\n", err)
		waitForExit()
		return
	}
	if !found {
		fmt.Printf("Boot file not found: %s\n", cfg.BootFilePath())
		fmt.Println("Create the file or update BOOT_FILE in watcher.ini")
		waitForExit()
		return
	}

	runner := functions.NewAppRunner(cfg)
	runner.SetupSignalHandler()

	fmt.Printf("Starting app in %s mode...\n", cfg.Mode)
	fmt.Printf("Watching for changes in: %s\n", cfg.ExeDir)

	var watchErr error
	switch cfg.Mode {
	case config.ModePolling:
		fmt.Println("Using polling-based watcher (Docker mode)")
		watchErr = functions.WatchPolling(cfg, runner)
	default:
		fmt.Println("Using native filesystem events")
		watchErr = functions.WatchRegular(cfg, runner)
	}

	if watchErr != nil {
		fmt.Printf("Error starting watcher: %v\n", watchErr)
		waitForExit()
		return
	}

	runner.RunApp()
	select {}
}

func waitForExit() {
	fmt.Println("Press ENTER to exit...")
	fmt.Scanln()
}
