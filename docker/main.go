package main

import (
	"fmt"
	"os"

	"github.com/alexanderthegreat96/nadeshot-watcher/functions"
	"github.com/common-nighthawk/go-figure"
	"github.com/radovskyb/watcher"
)

func main() {
	go functions.RunBot("main.py")
	w := watcher.New()

	// Set options
	w.SetMaxEvents(1)
	w.FilterOps(watcher.Write)

	currentPath, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current working directory:", err)
		fmt.Println("Press ENTER to exit...")
		fmt.Scanln()
		return
	}
	functions.WatcherWatchPath(w, currentPath)

	myFigure := figure.NewColorFigure("Nadeshot Watcher", "", "green", true)
	myFigure.Print()

	fmt.Println("\nStarting bot...")

	fmt.Println("Listening for file system changes in:", currentPath)

	select {}
}
