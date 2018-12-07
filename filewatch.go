package main

import (
	"log"

	"github.com/fsnotify/fsnotify"
)

func startFilewatcher(service *Service, path string) {
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
				log.Println("config changed, reloading")

				config, err := readConfig(path)
				if err != nil {
					log.Println("config error:", err)
					continue
				}
				if err := service.reload(config); err != nil {
					log.Println("reload error:", err)
					continue
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("watch error:", err)
		}
	}
}
