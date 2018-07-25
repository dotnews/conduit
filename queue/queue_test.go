package queue_test

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/dotnews/conduit/queue"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type HandlerMock struct {
	mock.Mock
	Done chan bool
}

func NewHandlerMock() *HandlerMock {
	return &HandlerMock{
		Done: make(chan bool),
	}
}

func (m *HandlerMock) handleFunc(message []byte) error {
	args := m.Called(message)
	close(m.Done)
	return args.Error(0)
}

func clean(c *redis.Client) {
	c.Del(c.Keys("test/queue/*").Val()...)
}

func TestMain(m *testing.M) {
	c := queue.NewClient()
	clean(c)
	code := m.Run()
	clean(c)
	os.Exit(code)
}

func TestPubSub(t *testing.T) {
	q := queue.New(1 * time.Millisecond)
	m := NewHandlerMock()

	q.Subscribe("test/queue/pubsub", m.handleFunc)
	m.AssertNotCalled(t, "handleFunc")

	m.On("handleFunc", []byte("message")).Return(nil)
	q.Publish("test/queue/pubsub", []byte("message"))

	select {
	case <-m.Done:
		m.AssertCalled(t, "handleFunc", []byte("message"))

		assert.Equal(t, int64(0), q.Client.LLen("test/queue/pubsub").Val())
		assert.Equal(t, int64(0), q.Client.LLen("test/queue/pubsub/proc").Val())

	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "Timeout")
	}
}

func TestMultipleHandlers(t *testing.T) {
	q := queue.New(1 * time.Millisecond)
	m1 := NewHandlerMock()
	m2 := NewHandlerMock()

	q.Subscribe("test/queue/multi", m1.handleFunc)
	q.Subscribe("test/queue/multi", m2.handleFunc)

	m1.AssertNotCalled(t, "handleFunc")
	m2.AssertNotCalled(t, "handleFunc")

	m1.On("handleFunc", []byte("message")).Return(nil)
	m2.On("handleFunc", []byte("message")).Return(nil)

	q.Publish("test/queue/multi", []byte("message"))

	select {
	case <-m1.Done:
		m1.AssertCalled(t, "handleFunc", []byte("message"))
		assert.Equal(t, int64(0), q.Client.LLen("test/queue/multi").Val())

	case <-m2.Done:
		m2.AssertCalled(t, "handleFunc", []byte("message"))
		assert.Equal(t, int64(0), q.Client.LLen("test/queue/multi").Val())
		assert.Equal(t, int64(0), q.Client.LLen("test/queue/multi/proc").Val())

	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "Timeout")
	}
}

func TestFailure(t *testing.T) {
	q := queue.New(1 * time.Millisecond)
	m := NewHandlerMock()

	q.Subscribe("test/queue/fail", m.handleFunc)
	m.AssertNotCalled(t, "handleFunc")

	m.On("handleFunc", []byte("message")).Return(errors.New(""))
	q.Publish("test/queue/fail", []byte("message"))

	select {
	case <-m.Done:
		m.AssertCalled(t, "handleFunc", []byte("message"))

		assert.Equal(t, int64(0), q.Client.LLen("test/queue/fail").Val())
		assert.Equal(t, int64(1), q.Client.LLen("test/queue/fail/proc").Val())

	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "Timeout")
	}
}
