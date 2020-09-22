package cmd2server

import (
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"
)

// Command ...
type Command struct {
	cmd    *exec.Cmd
	Reader *io.PipeReader
	Writer *io.PipeWriter
	Done   chan bool
}

// NewCommand ...
func NewCommand(args []string) *Command {
	reader, writer := io.Pipe()

	mw := io.MultiWriter(writer, os.Stdout)

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = mw
	cmd.Stderr = mw

	return &Command{
		cmd:    cmd,
		Reader: reader,
		Writer: writer,
		Done:   make(chan bool),
	}
}

// Start ...
func (c *Command) Start() error {
	err := c.cmd.Start()
	if err != nil {
		return err
	}

	go func() {
		c.cmd.Wait()
		c.Done <- true
	}()

	return nil
}

// Stop ...
func (c *Command) Stop() {
	log.Printf("killing PID %d with signal %s", c.cmd.Process.Pid, "SIGTERM")
	syscall.Kill(c.cmd.Process.Pid, syscall.SIGTERM)
	c.cmd.Wait()
	log.Printf("PID %d exited with %d", c.cmd.Process.Pid, c.cmd.ProcessState.ExitCode())
}

// Cleanup ...
func (c *Command) Cleanup() {
	c.Writer.Close()
	c.Reader.Close()
}
