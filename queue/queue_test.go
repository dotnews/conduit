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

var q = queue.New(1 * time.Nanosecond)

type HandlerMock struct {
	mock.Mock
}

func (m *HandlerMock) handle(message string) error {
	args := m.Called(message)
	return args.Error(0)
}

func TestMain(m *testing.M) {
	q.Client.Del("test", "test/proc")
	os.Exit(m.Run())
}

func TestPubSub(t *testing.T) {
	m := new(HandlerMock)

	q.Subscribe("test", m.handle)
	m.AssertNotCalled(t, "handle")

	m.On("handle", "message").Return(nil)
	q.Publish("test", "message")
	time.Sleep(100 * time.Millisecond)
	m.AssertCalled(t, "handle", "message")

	assert.Equal(t, int64(0), q.Client.LLen("test").Val())
	assert.Equal(t, int64(0), q.Client.LLen("test/proc").Val())
}

func TestPubSubFailure(t *testing.T) {
	m := new(HandlerMock)

	q.Subscribe("test", m.handle)
	m.AssertNotCalled(t, "handle")

	m.On("handle", "message").Return(errors.New(""))
	q.Publish("test", "message")
	time.Sleep(100 * time.Millisecond)
	m.AssertCalled(t, "handle", "message")

	assert.Equal(t, int64(0), q.Client.LLen("test").Val())
	assert.Equal(t, int64(1), q.Client.LLen("test/proc").Val())
}
