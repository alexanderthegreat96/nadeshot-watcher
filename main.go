package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alexanderthegreat96/nadeshot-watcher/config"
	"github.com/alexanderthegreat96/nadeshot-watcher/functions"
	"github.com/common-nighthawk/go-figure"
)

func printUsage() {
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		helpText := `
Usage: nadeshot-watcher [OPTIONS]

Python real-time application watcher and restarter.

Options:
  -h, --help          Show this screen.
  -p, --path <path>   Path to the directory containing watcher.ini.
                      (Defaults to the current executable directory)

Notes:
  If no path is provided, the program looks for 'watcher.ini' in the 
  same folder as the nadeshot-watcher binary.

		`
		fmt.Println(helpText)
	}
}

func main() {
	// divert execution
	// for standalone mode
	// we want to be able to make use of a different path
	// provide --path=whatever-path-you-want
	// or default which starts with this current exe
	myFigure := figure.NewColorFigure("nWatcher", "", "green", true)
	myFigure.Print()
	fmt.Println()

	args := os.Args[1:]

	var exeDir string
	if len(args) == 0 {
		// no path provided
		// use the path of the current executable
		exePath, err := os.Executable()
		if err != nil {
			fmt.Printf("Error getting executable path: %s\n", err)
			return
		}

		exeDir = filepath.Dir(exePath)

	} else {
		switch args[0] {
		case "--help", "-h":
			printUsage()
			return

		case "--path", "-p":
			if len(args) < 2 {
				fmt.Println("Error: --path requires a directory argument.")
				printUsage()
				os.Exit(1)
			}
			exeDir = args[1]

			if _, err := os.Stat(exeDir); os.IsNotExist(err) {
				fmt.Printf("Path: [%s] DOES NOT EXIST! Wrong path maybe?\n", exeDir)
				os.Exit(1)
			}

			fmt.Printf("Using Custom Watcher Path: %s\n", exeDir)

		default:
			fmt.Printf("Unknown argument: %s\n", args[0])
			printUsage()
			os.Exit(1)
		}
	}

	if exeDir == "" {
		fmt.Println("No executable path provided. Exiting...")
		return
	}

	cfg, err := config.LoadConfig(exeDir)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		waitForExit()
		return
	}

	fmt.Println(cfg)

	if !functions.IsPythonInstalled(cfg.PythonCommand) {
		fmt.Printf("Python interpreter '%s' not found.\n", cfg.PythonCommand)
		fmt.Printf("Please install Python or update the PYTHON setting in %s/watcher.ini\n", exeDir)
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
		fmt.Printf("Create the file or update BOOT_FILE in %s/watcher.ini\n", exeDir)
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
