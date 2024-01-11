package main

import (
	"fmt"
	"log"
	"os"

	"github.com/alexanderthegreat96/nadeshot-watcher/functions"
	"github.com/common-nighthawk/go-figure"
	"github.com/fsnotify/fsnotify"
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
			watcher, err := fsnotify.NewWatcher()
			if err != nil {
				log.Fatal(err)
				fmt.Println("Press ENTER to exit...")
				fmt.Scanln()
				return
			}
			defer watcher.Close()

			currentPath, err := os.Getwd()
			if err != nil {
				fmt.Println("Error getting current working directory:", err)
				fmt.Println("Press ENTER to exit...")
				fmt.Scanln()
				return
			}

			functions.WatchPath(watcher, currentPath, mainFile)

			fmt.Println("\nStarting app...")
			fmt.Println("Listening for file system changes in:", currentPath)

			functions.RunApp(mainFile)
			<-make(chan struct{})
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
