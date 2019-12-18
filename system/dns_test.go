package system

import (
	"testing"
)

func TestparseServerString(t *testing.T) {

	tables := []struct {
		x string
		n string
	}{
		{"127.0.0.1", "127.0.0.1:53"},
		{"127.0.0.1:53", "127.0.0.1:53"},
		{"127.0.0.1:8600", "127.0.0.1:8600"},
		{"1.1.1.1:53", "1.1.1.1:53"},
	}

	for _, table := range tables {
		output := parseServerString(table.x)
		if output != table.n {
			t.Errorf("parseServerString (%s) was incorrect, got: %s, want: %s.", table.x, output, table.n)
		}
	}
}
