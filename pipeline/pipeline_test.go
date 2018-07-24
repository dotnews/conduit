package pipeline_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang/glog"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"

	"github.com/dotnews/conduit/pipeline"
	"github.com/dotnews/conduit/queue"
)

const pipe = `---
id: test/pipeline/pipe

stages:
- process: echo foo
  subscribe: index
  publish: item`

const each = `---
id: test/pipeline/each

stages:
- process: 'echo "[{\"id\": 1}, {\"id\": 2}]"'
  subscribe: index
  publish: list
  pipe: each

- process: xargs echo
  subscribe: list
  publish: item`

const fail = `---
id: test/pipeline/fail

stages:
- process: foo
  subscribe: index
  publish: item`

const dir = `---
id: test/pipeline/dir

stages:
- process: pwd
  subscribe: index
  publish: item`

func newPipeline(root, s string, q *queue.Queue) *pipeline.Pipeline {
	var meta pipeline.Meta
	if err := yaml.Unmarshal([]byte(s), &meta); err != nil {
		glog.Fatalf("Failed parsing pipeline: %v\n%s", err, s)
	}
	return pipeline.New(root, &meta, q)
}

func clean() {
	q := queue.New(1 * time.Hour)
	q.Client.Del(q.Client.Keys("test/pipeline/*").Val()...)
}

func TestMain(m *testing.M) {
	clean()
	code := m.Run()
	clean()
	os.Exit(code)
}

func TestPipe(t *testing.T) {
	q := queue.New(1 * time.Millisecond)
	p := newPipeline(".", pipe, q)

	assert.Equal(t, "test/pipeline/pipe", p.Meta.ID)
	assert.Equal(t, 1, len(p.Meta.Stages))

	stage := p.Meta.Stages[0]

	assert.Equal(t, "echo foo", stage.Process)
	assert.Equal(t, "index", stage.Subscribe)
	assert.Equal(t, "item", stage.Publish)
	assert.Equal(t, pipeline.Pipe(""), stage.Pipe)

	p.Run()
	q.Publish("test/pipeline/pipe/index", []byte(""))
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, int64(0), q.Client.LLen("test/pipeline/pipe/index").Val())
	assert.Equal(t, int64(0), q.Client.LLen("test/pipeline/pipe/index/proc").Val())
	assert.Equal(t, int64(1), q.Client.LLen("test/pipeline/pipe/item").Val())
	assert.Equal(t, int64(0), q.Client.LLen("test/pipeline/pipe/item/proc").Val())

	assert.Equal(t, []string{"foo\n"}, q.Client.LRange("test/pipeline/pipe/item", 0, 0).Val())
}

func TestPipeEach(t *testing.T) {
	q := queue.New(1 * time.Millisecond)
	p := newPipeline(".", each, q)

	assert.Equal(t, "test/pipeline/each", p.Meta.ID)
	assert.Equal(t, 2, len(p.Meta.Stages))

	stage1 := p.Meta.Stages[0]
	stage2 := p.Meta.Stages[1]

	assert.Equal(t, `echo "[{\"id\": 1}, {\"id\": 2}]"`, stage1.Process)
	assert.Equal(t, "index", stage1.Subscribe)
	assert.Equal(t, "list", stage1.Publish)
	assert.Equal(t, pipeline.PipeEach, stage1.Pipe)

	assert.Equal(t, "xargs echo", stage2.Process)
	assert.Equal(t, "list", stage2.Subscribe)
	assert.Equal(t, "item", stage2.Publish)
	assert.Equal(t, pipeline.Pipe(""), stage2.Pipe)

	p.Run()
	q.Publish("test/pipeline/each/index", []byte(""))
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, int64(0), q.Client.LLen("test/pipeline/each/index").Val())
	assert.Equal(t, int64(0), q.Client.LLen("test/pipeline/each/index/proc").Val())
	assert.Equal(t, int64(0), q.Client.LLen("test/pipeline/each/list").Val())
	assert.Equal(t, int64(0), q.Client.LLen("test/pipeline/each/list/proc").Val())
	assert.Equal(t, int64(2), q.Client.LLen("test/pipeline/each/item").Val())
	assert.Equal(t, int64(0), q.Client.LLen("test/pipeline/each/item/proc").Val())

	assert.Equal(t, []string{"{id:2}\n"}, q.Client.LRange("test/pipeline/each/item", 0, 0).Val())
	assert.Equal(t, []string{"{id:1}\n"}, q.Client.LRange("test/pipeline/each/item", 1, 1).Val())
}

func TestProcessFail(t *testing.T) {
	q := queue.New(1 * time.Millisecond)
	p := newPipeline(".", fail, q)

	assert.Equal(t, "test/pipeline/fail", p.Meta.ID)
	assert.Equal(t, 1, len(p.Meta.Stages))

	stage := p.Meta.Stages[0]

	assert.Equal(t, "foo", stage.Process)
	assert.Equal(t, "index", stage.Subscribe)
	assert.Equal(t, "item", stage.Publish)
	assert.Equal(t, pipeline.Pipe(""), stage.Pipe)

	p.Run()
	q.Publish("test/pipeline/fail/index", []byte(""))
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, int64(0), q.Client.LLen("test/pipeline/fail/index").Val())
	assert.Equal(t, int64(1), q.Client.LLen("test/pipeline/fail/index/proc").Val())
	assert.Equal(t, int64(0), q.Client.LLen("test/pipeline/fail/item").Val())
	assert.Equal(t, int64(0), q.Client.LLen("test/pipeline/fail/item/proc").Val())
}

func TestDirectoryContext(t *testing.T) {
	tmp := os.TempDir()
	q := queue.New(1 * time.Millisecond)
	p := newPipeline(tmp, dir, q)

	assert.Equal(t, "test/pipeline/dir", p.Meta.ID)
	assert.Equal(t, 1, len(p.Meta.Stages))

	stage := p.Meta.Stages[0]

	assert.Equal(t, "pwd", stage.Process)
	assert.Equal(t, "index", stage.Subscribe)
	assert.Equal(t, "item", stage.Publish)
	assert.Equal(t, pipeline.Pipe(""), stage.Pipe)

	p.Run()
	q.Publish("test/pipeline/dir/index", []byte(""))
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, int64(0), q.Client.LLen("test/pipeline/dir/index").Val())
	assert.Equal(t, int64(0), q.Client.LLen("test/pipeline/dir/index/proc").Val())
	assert.Equal(t, int64(1), q.Client.LLen("test/pipeline/dir/item").Val())
	assert.Equal(t, int64(0), q.Client.LLen("test/pipeline/dir/item/proc").Val())

	assert.Contains(t, q.Client.LRange("test/pipeline/dir/item", 0, 0).Val()[0], strings.Trim(tmp, "/"))
}
