package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v3"
)

// FuzzValidate posts arbitrary bodies to /v1/validate without
// external-checks, so the parser never touches the network.
func FuzzValidate(f *testing.F) {
	fixtures, err := filepath.Glob("testdata/*.yml")
	if err != nil {
		f.Fatal(err)
	}

	for _, fixture := range fixtures {
		b, err := os.ReadFile(fixture)
		if err != nil {
			f.Fatal(err)
		}

		f.Add(b)
	}

	f.Fuzz(func(t *testing.T, body []byte) {
		req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewReader(body))

		res, err := app.Test(req, fiber.TestConfig{Timeout: 0})
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()

		if res.StatusCode == http.StatusInternalServerError {
			response, _ := io.ReadAll(res.Body)
			t.Fatalf("got 500 for body %q: %s", body, response)
		}
	})
}
