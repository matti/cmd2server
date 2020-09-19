package main

import (
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func main() {
	listen := os.Args[1]
	command := os.Args[2:]

	ln, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatal(err)
	}

	for {
		log.Printf("waiting for connection at %s", listen)

		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("accepted connection from %s", conn.RemoteAddr())
		handle(conn, command)
		conn.Close()
	}
}

func handle(conn net.Conn, command []string) {
	name := command[0]
	args := command[1:]

	log.Printf("exec %s with args %s", name, args)
	cmdReader, cmdWriter := io.Pipe()

	mw := io.MultiWriter(cmdWriter, os.Stdout)
	cmd := exec.Command(name, args...)
	cmd.Stdout = mw
	cmd.Stderr = mw
	err := cmd.Start()

	if err != nil {
		log.Fatalf("command start error %s", err)
	}

	go func() {
		cmd.Wait()
		// releases the blocking for loop
		cmdReader.Close()
	}()

	go func() {
		buf := make([]byte, 1)
		for {
			_, err := conn.Read(buf)
			if err != nil {
				conn.Close()
				return
			}
		}
	}()

	for {
		// TODO: buffer "clips" if outside of for ?
		buf := make([]byte, 4096)
		_, err := cmdReader.Read(buf)
		if err != nil {
			log.Printf("read err %s", err)
			break
		}
		conn.SetWriteDeadline(time.Now().Add(200 * time.Millisecond))
		_, err = conn.Write(buf)
		if err != nil {
			log.Printf("write err %s", err)
			break
		}
	}

	// can not use cmd.ProcessState because it's most likely nil in many cases until exited, so killing anyway
	log.Printf("killing PID %d with signal %s", cmd.Process.Pid, "SIGTERM")
	syscall.Kill(cmd.Process.Pid, syscall.SIGTERM)
	log.Printf("PID %d exited with %d", cmd.Process.Pid, cmd.ProcessState.ExitCode())

	conn.Close()
}
