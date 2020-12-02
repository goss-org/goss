package outputs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidFormat(t *testing.T) {
	if IsValidFormat("ne") {
		t.Fatal("'ne' should not be a valid output format")
	}

	if !IsValidFormat("json") {
		t.Fatal("'json' should be a valid output format")
	}
}

func TestOutputers(t *testing.T) {
	list := Outputers()
	assert.NotEmpty(t, list)
}

func TestGetOutputer(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		got, err := GetOutputer("rspecish")
		assert.NoError(t, err)
		assert.NotNil(t, got)
	})
	t.Run("not-valid", func(t *testing.T) {
		got, err := GetOutputer("gibberish")
		assert.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestOutputFormatOptions(t *testing.T) {
	list := FormatOptions()
	assert.NotEmpty(t, list)

	assert.Contains(t, list, foPerfData)
	assert.Contains(t, list, foPretty)
	assert.Contains(t, list, foVerbose)
	assert.Len(t, list, 3)
}
