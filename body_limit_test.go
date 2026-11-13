package main

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOversizedValidationBody(t *testing.T) {
	server := Setup()
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

	for _, method := range []string{http.MethodPost, "QUERY"} {
		t.Run(method, func(t *testing.T) {
			req, err := http.NewRequest(method, "http://"+listener.Addr().String()+"/v1/validate",
				strings.NewReader(strings.Repeat("x", 4*1024*1024+1)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/yaml")
			conn, err := net.DialTimeout("tcp", listener.Addr().String(), time.Second)
			require.NoError(t, err)
			defer conn.Close()
			require.NoError(t, conn.SetDeadline(time.Now().Add(5*time.Second)))
			// The server can respond and close before the upload finishes.
			// Read that response even if writing the remaining body fails.
			_ = req.Write(conn)
			res, err := http.ReadResponse(bufio.NewReader(conn), req)
			require.NoError(t, err)
			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, http.StatusRequestEntityTooLarge, res.StatusCode)
			assert.Equal(t, "application/problem+json", res.Header.Get("Content-Type"))
			assert.JSONEq(t, `{"title":"Request Entity Too Large","status":413}`, string(body))
		})
	}
}
