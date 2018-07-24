package queue_test

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/dotnews/conduit/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type HandlerMock struct {
	mock.Mock
}

func (m *HandlerMock) handleFunc(message []byte) error {
	args := m.Called(message)
	return args.Error(0)
}

func clean() {
	q := queue.New(1 * time.Hour)
	q.Client.Del(q.Client.Keys("test/queue/*").Val()...)
}

func TestMain(m *testing.M) {
	clean()
	code := m.Run()
	clean()
	os.Exit(code)
}

func TestPubSub(t *testing.T) {
	q := queue.New(1 * time.Millisecond)
	m := new(HandlerMock)

	q.Subscribe("test/queue/pubsub", m.handleFunc)
	m.AssertNotCalled(t, "handleFunc")

	m.On("handleFunc", []byte("message")).Return(nil)
	q.Publish("test/queue/pubsub", []byte("message"))
	time.Sleep(100 * time.Millisecond)
	m.AssertCalled(t, "handleFunc", []byte("message"))

	assert.Equal(t, int64(0), q.Client.LLen("test/queue/pubsub").Val())
	assert.Equal(t, int64(0), q.Client.LLen("test/queue/pubsub/proc").Val())
}

func TestPubSubFailure(t *testing.T) {
	q := queue.New(1 * time.Millisecond)
	m := new(HandlerMock)

	q.Subscribe("test/queue/fail", m.handleFunc)
	m.AssertNotCalled(t, "handleFunc")

	m.On("handleFunc", []byte("message")).Return(errors.New(""))
	q.Publish("test/queue/fail", []byte("message"))
	time.Sleep(100 * time.Millisecond)
	m.AssertCalled(t, "handleFunc", []byte("message"))

	assert.Equal(t, int64(0), q.Client.LLen("test/queue/fail").Val())
	assert.Equal(t, int64(1), q.Client.LLen("test/queue/fail/proc").Val())
}
