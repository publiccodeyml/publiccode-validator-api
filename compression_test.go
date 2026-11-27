package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCompressedBody(t *testing.T) {
	compress := func(body string) string {
		t.Helper()
		var buffer bytes.Buffer
		writer := gzip.NewWriter(&buffer)
		if _, err := writer.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		return buffer.String()
	}
	valid := loadTestdata(t, "valid.publiccode.yml")
	compressed := compress(valid)
	for _, method := range []string{http.MethodPost, "QUERY"} {
		for _, tc := range []struct {
			name, encoding, body string
			status               int
		}{
			{"malformed", "gzip", "not gzip", 400},
			{"truncated", "gzip", compressed[:len(compressed)-5], 400},
			{"oversized", "gzip", compress(strings.Repeat("a", 6_000_000)), 413},
			{"unsupported", "unknown", valid, 415},
			{"unimplemented", "compress", valid, 501},
			{"valid", "gzip", compressed, 200},
			{"stacked", "gzip, gzip", compress(compressed), 200},
		} {
			contentType := "application/problem+json"
			if tc.status == 200 {
				contentType = "application/json; charset=utf-8"
			}
			runTestCases(t, []TestCase{{
				description:         method + "/" + tc.name,
				query:               method + " /v1/validate",
				body:                tc.body,
				headers:             map[string][]string{"Content-Encoding": {tc.encoding}},
				expectedCode:        tc.status,
				expectedContentType: contentType,
				validateFunc: func(t *testing.T, response map[string]any) {
					if tc.status == 200 {
						assert.Equal(t, true, response["valid"])
						assert.Empty(t, response["results"])
						assert.NotEmpty(t, response["normalized"])
					} else {
						assert.Equal(t, float64(tc.status), response["status"])
						assert.NotEmpty(t, response["title"])
						for _, key := range []string{"valid", "results", "normalized"} {
							assert.NotContains(t, response, key, fmt.Sprint(response))
						}
					}
				},
			}})
		}
	}
}
