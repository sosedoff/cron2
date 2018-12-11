package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

var (
	replyInvalidCmd = []byte("err: invalid command")
	replyNoJob      = []byte("err: job name required")
	replyNotFound   = []byte("err: not found")
	replyOk         = []byte("ok: scheduled")
)

func startListener(service *Service, path string) error {
	if path == "" {
		return errors.New("Socket path is required")
	}

	listener, err := net.Listen("unix", path)
	if err != nil {
		return err
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
			chunks := strings.Split(input, " ")

			switch chunks[0] {
			case "run":
				if len(chunks) < 2 {
					conn.Write(replyNoJob)
					return
				}
				jobConfig := service.config.findJob(strings.Join(chunks[1:], " "))
				if jobConfig == nil {
					conn.Write(replyNotFound)
					return
				}
				job := Job{config: jobConfig}
				go job.Run()
				conn.Write(replyOk)
			case "list":
				names := []string{}
				for _, j := range service.config.Jobs {
					names = append(names, fmt.Sprintf("%s: %s", j.Name, j.state()))
				}
				conn.Write([]byte(strings.Join(names, "\n")))
			default:
				conn.Write(replyInvalidCmd)
			}
		}()
	}

	return nil
}
