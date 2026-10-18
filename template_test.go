package goss

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateGlob(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"b.xml", "a.xml", "c.txt"} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("<a/>"), 0o644))
	}

	filter, err := NewTemplateFilter(nil, `{"dir": "`+dir+`"}`)
	require.NoError(t, err)

	out, err := filter([]byte(`{{ range glob (printf "%s/*.xml" .Vars.dir) }}{{ base . }} {{ end }}`))
	require.NoError(t, err)
	assert.Equal(t, "a.xml b.xml ", string(out))

	out, err = filter([]byte(`{{ range glob (printf "%s/*.none" .Vars.dir) }}{{ . }}{{ else }}empty{{ end }}`))
	require.NoError(t, err)
	assert.Equal(t, "empty", string(out))
}
