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
	// Create new watcher.
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

	functions.WatchPath(watcher, currentPath)

	myFigure := figure.NewColorFigure("Nadeshot Watcher", "", "green", true)
	myFigure.Print()

	fmt.Println("\nStarting bot...")

	fmt.Println("Listening for file system changes in:", currentPath)

	functions.RunBot("main.py")
	<-make(chan struct{})
}
