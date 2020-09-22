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

	clientLostChan := make(chan bool)
	processExitChan := make(chan bool)

	go func(done chan bool) {
		buf := make([]byte, 1)
		for {
			_, err := conn.Read(buf)
			if err != nil {
				break
			}
		}

		done <- true
	}(clientLostChan)

	go func() {
		for {
			// TODO: buffer "clips" if outside of for ?
			buf := make([]byte, 4096)
			_, err := cmdReader.Read(buf)
			if err != nil {
				break
			}

			conn.SetWriteDeadline(time.Now().Add(200 * time.Millisecond))
			_, err = conn.Write(buf)
			if err != nil {
				break
			}
		}

	}()

	go func(done chan bool) {
		cmd.Wait()
		processExitChan <- true
	}(processExitChan)

	select {
	case <-processExitChan:
		log.Printf("process exited")
	case <-clientLostChan:
		log.Printf("client lost, killing PID %d with signal %s", cmd.Process.Pid, "SIGTERM")
		syscall.Kill(cmd.Process.Pid, syscall.SIGTERM)
	}

	cmd.Wait()
	log.Printf("PID %d exited with %d", cmd.Process.Pid, cmd.ProcessState.ExitCode())

	cmdReader.Close()
	cmdWriter.Close()
	conn.Close()
}
