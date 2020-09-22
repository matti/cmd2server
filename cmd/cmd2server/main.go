package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	cmd2server "github.com/matti/cmd2server/internal"
)

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)
	go func() {
		s := <-sigChan
		log.Printf("got signal %d, exit", s)
		os.Exit(0)
	}()

	listen := os.Args[1]

	ln, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatal(err)
	}

	for {
		log.Printf("PID %d waiting for connection at %s", os.Getpid(), listen)

		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("accepted connection from %s", conn.RemoteAddr())

		command := cmd2server.NewCommand(os.Args[2:])
		handle(conn, command)
		command.Cleanup()
		conn.Close()
	}
}

func handle(conn net.Conn, command *cmd2server.Command) {
	clientLostChan := make(chan bool)
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

	err := command.Start()
	if err != nil {
		log.Fatalf("command start error %s", err)
	}

	go func() {
		for {
			// TODO: buffer "clips" if outside of for ?
			buf := make([]byte, 4096)
			_, err := command.Reader.Read(buf)
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

	select {
	case <-command.Done:
		log.Println("command done")
	case <-clientLostChan:
		log.Println("client lost")
		command.Stop()
	}
}
