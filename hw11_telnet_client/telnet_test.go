package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClient_Connect(t *testing.T) {
	server, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer server.Close()

	go func() {
		conn, _ := server.Accept()
		if conn != nil {
			conn.Close()
		}
	}()

	addr := server.Addr().(*net.TCPAddr)
	// Исправлено: используем fmt.Sprintf для формирования адреса
	address := fmt.Sprintf("127.0.0.1:%d", addr.Port)

	client := New(address, 1*time.Second, io.NopCloser(&bytes.Buffer{}), &bytes.Buffer{})
	err = client.Connect()
	require.NoError(t, err)
}

func TestClient_SendEOF(t *testing.T) {
	server, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer server.Close()

	done := make(chan struct{})
	go func() {
		conn, _ := server.Accept()
		if conn != nil {
			buf := make([]byte, 1024)
			conn.Read(buf) // Читаем "hello"
			conn.Close()
		}
		close(done)
	}()

	addr := server.Addr().(*net.TCPAddr)
	address := fmt.Sprintf("127.0.0.1:%d", addr.Port)

	input := bytes.NewBufferString("hello")
	client := New(address, 1*time.Second, io.NopCloser(input), &bytes.Buffer{})

	require.NoError(t, client.Connect())
	err = client.Send()
	<-done
	// Тест ожидает nil, так как ваш Send() теперь возвращает nil при EOF (для успеха основного пайплайна)
	// Если вы хотите проверить, что Send завершился, достаточно require.NoError
	require.NoError(t, err)
}

func TestClient_Receive(t *testing.T) {
	server, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer server.Close()

	go func() {
		conn, _ := server.Accept()
		if conn != nil {
			conn.Write([]byte("world\n"))
			conn.Close()
		}
	}()

	addr := server.Addr().(*net.TCPAddr)
	address := fmt.Sprintf("127.0.0.1:%d", addr.Port)

	output := &bytes.Buffer{}
	client := New(address, 1*time.Second, io.NopCloser(&bytes.Buffer{}), output)

	require.NoError(t, client.Connect())
	err = client.Receive()

	// Receive() в вашей реализации возвращает nil при EOF, поэтому проверяем на NoError
	require.NoError(t, err)
	require.Equal(t, "world\n", output.String())
}
