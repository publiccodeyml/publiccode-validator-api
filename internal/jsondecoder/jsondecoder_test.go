package jsondecoder_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/italia/publiccode-validator-api/internal/jsondecoder"
)

type target struct {
	Name string `json:"name"`
}

func TestUnmarshalDisallowUnknownFieldsAcceptsKnownField(t *testing.T) {
	var v target

	err := jsondecoder.UnmarshalDisallowUnknownFields([]byte(`{"name": "foo"}`), &v)

	assert.NoError(t, err)
	assert.Equal(t, "foo", v.Name)
}

func TestUnmarshalDisallowUnknownFieldsRejectsUnknownField(t *testing.T) {
	tests := []struct {
		description string
		data        string
	}{
		{"unknown field alone", `{"unknown": true}`},
		{"unknown field after a known one", `{"name": "foo", "unknown": true}`},
		{"field differing from a known one", `{"name_": "foo"}`},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			var v target

			err := jsondecoder.UnmarshalDisallowUnknownFields([]byte(test.data), &v)

			assert.ErrorIs(t, err, jsondecoder.ErrUnknownField)
		})
	}
}
