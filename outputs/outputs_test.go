package outputs

import (
	"testing"
)

func TestIsValidFormat(t *testing.T) {
	if IsValidFormat("ne") {
		t.Fatal("'ne' should not be a valid output format")
	}

	if !IsValidFormat("json") {
		t.Fatal("'json' should be a valid output format")
	}
}

func TestIsValidFormatOption(t *testing.T) {
	if IsValidFormatOption("ne") {
		t.Fatal("'ne' should not be a valid output format option")
	}

	if !IsValidFormatOption("verbose") {
		t.Fatal("'verbose' should be a valid output format option")
	}
}
