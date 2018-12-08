package main

import (
	"fmt"
	"net"
)

func listCurrentJobs(socketPath string) error {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.Write([]byte("list")); err != nil {
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
