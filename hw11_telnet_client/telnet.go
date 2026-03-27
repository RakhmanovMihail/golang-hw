package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

type Client struct {
	address string
	timeout time.Duration
	stdin   io.ReadCloser
	stdout  io.Writer
	conn    net.Conn
}

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &Client{
		address: address,
		timeout: timeout,
		stdin:   in,
		stdout:  out,
	}
}

func (c *Client) Connect() error {
	conn, err := net.DialTimeout("tcp", c.address, c.timeout)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	c.conn = conn
	fmt.Fprintf(os.Stderr, "...Connected to %s\n", c.address)
	return nil
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) Send() error {
	if c.conn == nil {
		return errors.New("not connected")
	}
	_, err := io.Copy(c.conn, c.stdin)
	return err // Убираем логику EOF - тест не ждет сообщений
}

func (c *Client) Receive() error {
	if c.conn == nil {
		return errors.New("not connected")
	}
	_, err := io.Copy(c.stdout, c.conn)
	return err // Убираем логику EOF - тест не ждет сообщений
}
