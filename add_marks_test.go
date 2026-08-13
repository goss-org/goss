package goss

import (
	"testing"

	"github.com/goss-org/goss/resource"
	"github.com/stretchr/testify/assert"
)

func TestApplyMarksIfUnset(t *testing.T) {
	t.Parallel()

	t.Run("nil resource is a no-op", func(t *testing.T) {
		t.Parallel()
		// Must not panic.
		applyMarksIfUnset(nil, []string{"critical"})
	})

	t.Run("empty marks leaves resource untouched", func(t *testing.T) {
		t.Parallel()
		r := &resource.Addr{}
		applyMarksIfUnset(r, nil)
		assert.Empty(t, r.GetMarks())

		applyMarksIfUnset(r, []string{})
		assert.Empty(t, r.GetMarks())
	})

	t.Run("applies marks when resource has none", func(t *testing.T) {
		t.Parallel()
		r := &resource.Addr{}
		applyMarksIfUnset(r, []string{"critical", "fast"})
		assert.Equal(t, []string{"critical", "fast"}, r.GetMarks())
	})

	t.Run("preserves existing marks", func(t *testing.T) {
		t.Parallel()
		r := &resource.Addr{Marks: []string{"existing"}}
		applyMarksIfUnset(r, []string{"critical"})
		assert.Equal(t, []string{"existing"}, r.GetMarks())
	})

	t.Run("copies input slice to avoid aliasing", func(t *testing.T) {
		t.Parallel()
		r1 := &resource.Addr{}
		r2 := &resource.Addr{}
		input := []string{"a", "b"}

		applyMarksIfUnset(r1, input)
		applyMarksIfUnset(r2, input)

		// Mutate r1's marks; r2 must not see the change.
		r1.GetMarks()[0] = "mutated"
		assert.Equal(t, []string{"a", "b"}, r2.GetMarks(),
			"applyMarksIfUnset must copy the slice so resources don't share backing arrays")
	})

	t.Run("works across multiple resource types", func(t *testing.T) {
		t.Parallel()
		cases := []resource.ResourceRead{
			&resource.Addr{},
			&resource.Command{},
			&resource.File{},
			&resource.Group{},
			&resource.HTTP{},
			&resource.Package{},
			&resource.Port{},
			&resource.Process{},
			&resource.Service{},
			&resource.User{},
			&resource.DNS{},
			&resource.KernelParam{},
			&resource.Mount{},
			&resource.Interface{},
			&resource.Gossfile{},
		}
		for _, r := range cases {
			applyMarksIfUnset(r, []string{"production"})
			assert.Equal(t, []string{"production"}, r.GetMarks())
		}
	})
}
