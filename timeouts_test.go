package main

import (
	"io"
	"net"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

func TestStalledRequests(t *testing.T) {
	// Setup writes package globals in libpubliccode, so the subtests share
	// one server instead of each building their own in parallel.
	server := Setup()
	config := server.Config()
	require.Equal(t, 30*time.Second, config.ReadTimeout)
	require.Equal(t, 30*time.Second, config.WriteTimeout)
	require.Equal(t, 60*time.Second, config.IdleTimeout)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	done := make(chan error, 1)
	go func() {
		done <- server.Listener(listener, fiber.ListenConfig{DisableStartupMessage: true})
	}()
	t.Cleanup(func() {
		require.NoError(t, server.Shutdown())
		require.NoError(t, <-done)
	})

	for name, request := range map[string]string{
		"headers": "GET /v1/status HTTP/1.1\r\nHost: localhost\r\n",
		"body":    "POST /v1/validate HTTP/1.1\r\nHost: localhost\r\nContent-Length: 100\r\n\r\npubliccodeYmlVersion:",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			conn, err := net.DialTimeout("tcp", listener.Addr().String(), time.Second)
			require.NoError(t, err)
			defer conn.Close()
			require.NoError(t, conn.SetDeadline(time.Now().Add(config.ReadTimeout+5*time.Second)))
			_, err = io.WriteString(conn, request)
			require.NoError(t, err)

			// The server may send a timeout response before closing the connection.
			_, err = io.Copy(io.Discard, conn)
			require.NoError(t, err, "server must close the connection before the client deadline")
		})
	}
}
