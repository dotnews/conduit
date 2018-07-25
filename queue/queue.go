package queue

import (
	"os"
	"time"

	"github.com/golang/glog"

	"github.com/go-redis/redis"
)

const (
	procSuffix  = "/proc"
	redisNilErr = "redis: nil"
)

// Queue backed by redis
type Queue struct {
	Interval time.Duration
	Client   *redis.Client
	Handlers map[string][]HandleFunc
}

// HandleFunc for event messages
type HandleFunc func(message []byte) error

// New creates a new queue
func New(interval time.Duration) *Queue {
	q := &Queue{
		Interval: interval,
		Client:   NewClient(),
		Handlers: map[string][]HandleFunc{},
	}
	go q.listen()
	return q
}

// NewClient creates a new redis client
func NewClient() *redis.Client {
	c := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})
	if err := c.Ping().Err(); err != nil {
		glog.Fatalf("Failed connecting to redis: %v", err)
	}
	return c
}

// Publish message
func (q *Queue) Publish(event string, message []byte) error {
	if err := q.Client.LPush(event, message).Err(); err != nil {
		glog.Errorf(
			"Failed publishing message; event: %s, error: %v, message: %s",
			event,
			err,
			string(message),
		)
		return err
	}
	return nil
}

// Subscribe to event
func (q *Queue) Subscribe(event string, handleFunc HandleFunc) {
	q.Handlers[event] = append(q.Handlers[event], handleFunc)
}

func (q *Queue) listen() {
	ticker := time.NewTicker(q.Interval)
	for range ticker.C {
		for event, handleFuncs := range q.Handlers {
			q.proc(event, handleFuncs)
		}
	}
}

func (q *Queue) proc(event string, handleFuncs []HandleFunc) {
	proc := event + procSuffix
	message, err := q.Client.RPopLPush(event, proc).Bytes()
	if err != nil && err.Error() == redisNilErr {
		return
	}

	if err != nil {
		glog.Errorf(
			"Failed reading message; event: %s, error: %v",
			event,
			err,
		)
		return
	}

	failed := false
	for _, handleFunc := range handleFuncs {
		if err = handleFunc(message); err != nil {
			glog.Errorf(
				"Failed processing message; proc: %s, error: %v, message: %s",
				proc,
				err,
				string(message),
			)
			failed = true
		}
	}

	if failed {
		return
	}

	if err = q.Client.LRem(proc, 1, message).Err(); err != nil {
		glog.Errorf("Failed acknowledging message; proc: %s, error: %v, message: %s",
			proc,
			err,
			string(message),
		)
	}
}
