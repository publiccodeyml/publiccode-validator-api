package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

func TestShutdownProcess(t *testing.T) {
	if os.Getenv("TEST_SHUTDOWN_PROCESS") != "1" {
		return
	}

	server := Setup()
	server.Hooks().OnListen(func(data fiber.ListenData) error {
		fmt.Println("test:address:http://" + data.Host + ":" + data.Port)
		return nil
	})
	server.Hooks().OnPreShutdown(func() error {
		fmt.Println("test:shutdown")
		return nil
	})
	server.Get("/slow", func(ctx fiber.Ctx) error {
		fmt.Println("test:started")
		var release [1]byte
		if _, err := os.Stdin.Read(release[:]); err != nil {
			return err
		}
		return ctx.SendString("completed")
	})
	require.NoError(t, listen(server, "127.0.0.1:0"))
}

func TestShutdownFinishesRequest(t *testing.T) {
	for name, sig := range map[string]os.Signal{"SIGTERM": syscall.SIGTERM, "SIGINT": os.Interrupt} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestShutdownProcess$")
			cmd.Env = append(os.Environ(), "TEST_SHUTDOWN_PROCESS=1")
			cmd.Stderr = os.Stderr
			stdout, err := cmd.StdoutPipe()
			require.NoError(t, err)
			stdin, err := cmd.StdinPipe()
			require.NoError(t, err)
			require.NoError(t, cmd.Start())
			done := make(chan struct{})
			var exitErr error
			t.Cleanup(func() {
				_ = stdin.Close()
				_ = cmd.Process.Kill()
				<-done
			})

			events := make(chan string, 8)
			go func() {
				defer close(events)
				scanner := bufio.NewScanner(stdout)
				for scanner.Scan() {
					if strings.HasPrefix(scanner.Text(), "test:") {
						events <- scanner.Text()
					}
				}
			}()
			go func() {
				exitErr = cmd.Wait()
				close(done)
			}()
			nextEvent := func() string {
				t.Helper()
				select {
				case event, ok := <-events:
					require.True(t, ok, "server exited before the expected event")
					return event
				case <-ctx.Done():
					t.Fatal("timed out waiting for the server")
					return ""
				}
			}

			address := nextEvent()
			require.True(t, strings.HasPrefix(address, "test:address:"))
			request, err := http.NewRequestWithContext(ctx, http.MethodGet,
				strings.TrimPrefix(address, "test:address:")+"/slow", nil)
			require.NoError(t, err)
			response := make(chan *http.Response, 1)
			requestErr := make(chan error, 1)
			go func() {
				res, err := http.DefaultClient.Do(request)
				response <- res
				requestErr <- err
			}()

			require.Equal(t, "test:started", nextEvent())
			require.NoError(t, cmd.Process.Signal(sig))
			require.Equal(t, "test:shutdown", nextEvent())
			// Release the handler only after the real signal starts shutdown.
			_, err = stdin.Write([]byte{1})
			require.NoError(t, err)
			require.NoError(t, <-requestErr)
			res := <-response
			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, res.StatusCode)
			require.Equal(t, "completed", string(body))
			<-done
			require.NoError(t, exitErr, "shutdown must exit successfully")
		})
	}
}
