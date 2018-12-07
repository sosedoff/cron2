package main

import (
	"log"

	"github.com/fsnotify/fsnotify"
)

// TODO: actually reload config on any changes
func startFilewatcher(path string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println("cant start filewatcher:", err)
		return
	}
	defer watcher.Close()

	err = watcher.Add(path)
	if err != nil {
		return
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("modified file:", event.Name)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("watch error:", err)
		}
	}
}
