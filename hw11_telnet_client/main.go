package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	timeout := flag.Duration("timeout", 10*time.Second, "connection timeout")
	flag.Parse()

	if len(flag.Args()) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go-telnet [--timeout=10s] host port")
		os.Exit(2)
	}

	address := net.JoinHostPort(flag.Args()[0], flag.Args()[1])

	client := NewTelnetClient(address, *timeout, os.Stdin, os.Stdout)
	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "connect failed: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	done := make(chan error, 2)

	go func() { done <- client.Send() }()
	go func() { done <- client.Receive() }()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	select {
	case _ = <-done:
		<-done // Ждем вторую горутину
	case <-sigs:
		fmt.Fprintln(os.Stderr, "\n...Interrupted")
	}
}
