package pipeline_test

import (
	"testing"
	"time"

	"github.com/dotnews/conduit/queue"
	"github.com/stretchr/testify/assert"

	"github.com/dotnews/conduit/pipeline"
)

var q = queue.New(1 * time.Millisecond)
var p = pipeline.New("sample.yml", q, "")

func TestNew(t *testing.T) {
	assert.Equal(t, "sample", p.Meta.ID)
	assert.Equal(t, 2, len(p.Meta.Stages))

	s1 := p.Meta.Stages[0]
	s2 := p.Meta.Stages[1]

	assert.Equal(t, `echo "[{\"id\": 1}, {\"id\": 2}]"`, s1.Process)
	assert.Equal(t, "index", s1.Subscribe)
	assert.Equal(t, "list", s1.Publish)
	assert.Equal(t, pipeline.PipeEach, s1.Pipe)

	assert.Equal(t, "xargs echo", s2.Process)
	assert.Equal(t, "list", s2.Subscribe)
	assert.Equal(t, "item", s2.Publish)
	assert.Equal(t, pipeline.Pipe(""), s2.Pipe)
}
