package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStore(t *testing.T) {
	t.Parallel()

	s := New[string, int](map[string]int{
		"bar": 2,
	})

	val, ok := s.Get("bar")
	assert.Equal(t, 2, val)
	assert.True(t, ok)

	_, ok = s.Get("foo")
	assert.False(t, ok)

	s.Set("foo", 1)
	val, ok = s.Get("foo")
	assert.Equal(t, 1, val)
	assert.True(t, ok)

	assert.EqualValues(t, map[string]int{"foo": 1, "bar": 2}, s.Data())
}
