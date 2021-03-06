package pipeline

import (
	"encoding/json"
	"fmt"

	"github.com/golang/glog"

	"github.com/dotnews/conduit/os"
	"github.com/dotnews/conduit/queue"
)

// PipeEach configures the stage to handle output as JSON array
// piping each array element independently into the next stage
const PipeEach = Pipe("each")

// Pipeline metadata and runtime backed by queue
type Pipeline struct {
	Root  string
	Meta  *Meta
	Queue *queue.Queue
}

// Meta data definition of a pipeline
type Meta struct {
	ID     string
	Stages []Stage
}

// Stage entry in pipeline stage list
type Stage struct {
	Process   string
	Subscribe string
	Publish   string
	Pipe      Pipe
}

// PipeArray structure for parsing
type PipeArray []interface{}

// Pipe output mode
type Pipe string

// New creates a new pipeline
func New(root string, meta *Meta, q *queue.Queue) *Pipeline {
	return &Pipeline{
		Root:  root,
		Meta:  meta,
		Queue: q,
	}
}

// Run pipeline
func (p *Pipeline) Run() {
	for _, stage := range p.Meta.Stages {
		stage := stage // closure-safety
		event := p.getEvent(stage.Subscribe)
		glog.Infof("Subscribing to event: %s", event)

		p.Queue.Subscribe(event, func(message []byte) error {
			out, err := os.Run(stage.Process, p.Root, message)
			if err != nil {
				glog.Errorf(
					"Failed processing message; event: %s, error: %v, message: %s",
					event,
					err,
					string(message),
				)
				return err
			}

			next := p.getEvent(stage.Publish)
			count, err := p.pipe(next, stage.Pipe, out)
			if err != nil {
				glog.Errorf(
					"Failed piping message; event: %s, next: %s, error: %v, message: %s",
					event,
					next,
					err,
					string(message),
				)
				return err
			}

			glog.Infof(
				"Pipeline: %s, sub: %s, pub: %s, piped: %d",
				p.Meta.ID,
				stage.Subscribe,
				stage.Publish,
				count,
			)

			return nil
		})
	}
}

// Pipe message(s) into next stage
func (p *Pipeline) pipe(event string, pipe Pipe, message []byte) (int, error) {
	if pipe == PipeEach {
		return p.pipeEach(event, message)
	}

	if err := p.Queue.Publish(event, message); err != nil {
		glog.Errorf(
			"Failed publishing message; event: %s, error: %v, message: %s",
			event,
			err,
			string(message),
		)
		return 0, err
	}

	return 1, nil
}

// Pipe each message into next stage
func (p *Pipeline) pipeEach(event string, message []byte) (int, error) {
	var pipeArray PipeArray
	if err := json.Unmarshal(message, &pipeArray); err != nil {
		glog.Errorf(
			"Failed parsing message; event: %s, error: %v, message: %s",
			event,
			err,
			string(message),
		)
		return 0, err
	}

	for i, el := range pipeArray {
		m, err := json.Marshal(el)
		if err != nil {
			glog.Errorf(
				"Failed parsing message %d/%d; event: %s, error: %v, message: %s",
				i+1,
				len(pipeArray),
				event,
				err,
				string(m),
			)
			return 0, err
		}

		if err := p.Queue.Publish(event, m); err != nil {
			glog.Errorf(
				"Failed publishing message %d/%d; event: %s, error: %v, message: %s",
				i+1,
				len(pipeArray),
				event,
				err,
				string(m),
			)
			return 0, err
		}
	}

	return len(pipeArray), nil
}

// Get event name prefixed with pipeline ID
func (p *Pipeline) getEvent(suffix string) string {
	return fmt.Sprintf("%s/%s", p.Meta.ID, suffix)
}
