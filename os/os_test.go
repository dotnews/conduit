package os_test

import (
	"strings"
	"testing"

	"github.com/dotnews/conduit/os"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	out, err := os.Run("ls", ".", []byte{})
	assert.Nil(t, err)
	assert.True(t, strings.Contains(string(out), "os_test.go"))
}

func TestRunStdIn(t *testing.T) {
	out, err := os.Run("xargs echo", ".", []byte("foo"))
	assert.Nil(t, err)
	assert.True(t, strings.Contains(string(out), "foo"))
}
