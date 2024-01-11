package main

import (
	"fmt"
	"os"

	"github.com/alexanderthegreat96/nadeshot-watcher/functions"
	"github.com/common-nighthawk/go-figure"
	"github.com/radovskyb/watcher"
)

func main() {
	myFigure := figure.NewColorFigure("Nadeshot Watcher", "", "green", true)
	myFigure.Print()
	fmt.Println()

	if functions.IsPythonInstalled() {

		mainFile := "main.py"
		customFile, isDefined := functions.CustomBootFileDefined()

		if isDefined {
			mainFile = customFile
		}

		found, notFoundMain := functions.FileExists(mainFile)

		if notFoundMain != nil {
			fmt.Printf("Unable to run app entrypoint: %s\nError: %s\n", mainFile, notFoundMain.Error())
			fmt.Println("Press ENTER to exit...")
			fmt.Scanln()
			return
		}

		if found {
			currentPath, err := os.Getwd()
			if err != nil {
				fmt.Println("Error getting current working directory:", err)
				fmt.Println("Press ENTER to exit...")
				fmt.Scanln()
				return
			}

			go functions.RunApp(mainFile)
			w := watcher.New()

			w.SetMaxEvents(1)
			w.FilterOps(watcher.Write)

			functions.WatcherWatchPath(w, currentPath, mainFile)

			fmt.Println("\nStarting app...")
			fmt.Println("Listening for file system changes in:", currentPath)

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
