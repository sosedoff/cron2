package main

import (
	"log"
	"net"
	"os"
	"strings"
)

var (
	replyNoJob    = []byte("err: job name required")
	replyNotFound = []byte("err: not found")
	replyOk       = []byte("ok: scheduled")
)

func startListener(config *Config, path string) {
	if path == "" {
		return
	}

	os.Remove(path)

	listener, err := net.Listen("unix", path)
	if err != nil {
		log.Println("cant start server:", err)
		return
	}
	defer func() {
		listener.Close()
		os.Remove(path)
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("cant accept connection:", err)
			continue
		}

		go func() {
			defer conn.Close()

			buf := make([]byte, 256)

			n, err := conn.Read(buf)
			if err != nil {
				log.Println("read error:", err)
				return
			}

			input := strings.TrimSpace(string(buf[0:n]))
			if input == "" {
				conn.Write(replyNoJob)
				return
			}

			jobConfig := config.findJob(input)
			if jobConfig == nil {
				conn.Write(replyNotFound)
				return
			}

			job := Job{config: jobConfig}
			go job.Run()

			conn.Write(replyOk)
		}()
	}
}
