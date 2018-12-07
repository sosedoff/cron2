package main

import (
	"fmt"
	"net"
)

func triggerJob(socketPath string, job string) error {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.Write([]byte(job)); err != nil {
		return err
	}

	buff := make([]byte, 1024)

	n, err := conn.Read(buff)
	if err != nil {
		return err
	}

	fmt.Println(string(buff[0:n]))
	return nil
}
