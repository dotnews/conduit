package queue

import (
	"os"
	"time"

	"github.com/golang/glog"

	"github.com/go-redis/redis"
)

const procSuffix = "/proc"

// Queue backed by redis
type Queue struct {
	Interval    time.Duration
	Client      *redis.Client
	Subscribers map[string]HandleFunc
}

// HandleFunc for event messages
type HandleFunc func(message string) error

// New creates a new queue
func New(interval time.Duration) *Queue {
	q := &Queue{
		Interval: interval,
		Client: redis.NewClient(&redis.Options{
			Addr: os.Getenv("REDIS_ADDR"),
		}),
		Subscribers: map[string]HandleFunc{},
	}
	go q.listen()
	return q
}

// Publish event messages
func (q *Queue) Publish(event string, messages ...string) error {
	err := q.Client.LPush(event, messages).Err()
	if err != nil {
		glog.Errorf(
			"Failed publishing to event: %s, error: %v, messages: %+v",
			event,
			err,
			messages,
		)
	}
	return err
}

// Subscribe to event messages
func (q *Queue) Subscribe(event string, handle HandleFunc) {
	q.Subscribers[event] = handle
}

func (q *Queue) listen() {
	ticker := time.NewTicker(q.Interval)

	for range ticker.C {
		for event, handler := range q.Subscribers {
			q.proc(event, handler)
		}
	}
}

func (q *Queue) proc(event string, handle HandleFunc) {
	proc := event + procSuffix
	message, err := q.Client.RPopLPush(event, proc).Bytes()

	if err != nil {
		glog.Errorf(
			"Failed reading message; event: %s, error: %v",
			event,
			err,
		)
		return
	}

	if message == nil {
		return
	}

	err = handle(string(message))

	if err != nil {
		glog.Errorf(
			"Failed handling message; proc: %s, error: %v, message: %s",
			proc,
			err,
			string(message),
		)
		return
	}

	err = q.Client.LRem(proc, 1, message).Err()

	if err != nil {
		glog.Errorf(
			"Failed acknowledging message; proc: %s, error: %v, message: %s",
			proc,
			err,
			string(message),
		)
	}
}
