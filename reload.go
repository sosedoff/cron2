package main

import (
	"fmt"
	"net"
)

func reloadConfig(socketPath string) error {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.Write([]byte("reload")); err != nil {
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
