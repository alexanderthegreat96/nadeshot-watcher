package main

import (
	"fmt"
	"os"

	"github.com/alexanderthegreat96/nadeshot-watcher/functions"
	"github.com/common-nighthawk/go-figure"
	"github.com/radovskyb/watcher"

	"path/filepath"
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

	if functions.IsPythonInstalled() {

		mainFile := "main.py"
		customFile, isDefined := functions.CustomBootFileDefined()

		if isDefined {
			mainFile = customFile
		}
		mainFilePath := filepath.Join(exeDir, mainFile)
		found, notFoundMain := functions.FileExists(mainFilePath)

		if notFoundMain != nil {
			fmt.Printf("Unable to run app entrypoint: %s\nError: %s\n", mainFile, notFoundMain.Error())
			fmt.Println("Press ENTER to exit...")
			fmt.Scanln()
			return
		}

		if found {
			go functions.RunApp(mainFile)
			w := watcher.New()

			w.SetMaxEvents(1)
			w.FilterOps(watcher.Write)

			functions.WatcherWatchPath(w, exeDir, mainFile)

			fmt.Println("\nStarting app...")
			fmt.Println("Listening for file system changes in:", exeDir)

			select {}
		} else {
			fmt.Println("App entrypoint: " + mainFile + " was not found. Canceling...")
			fmt.Println("Press ENTER to exit...")
			fmt.Scanln()
		}
	} else {
		fmt.Println("No python environment found. Please install it and re-run the program.")
		fmt.Println("Press ENTER to exit...")
		fmt.Scanln()
	}

}
