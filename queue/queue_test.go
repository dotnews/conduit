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

var q = queue.New(1 * time.Millisecond)

type HandlerMock struct {
	mock.Mock
}

func (m *HandlerMock) handleFunc(message []byte) error {
	args := m.Called(message)
	return args.Error(0)
}

func TestMain(m *testing.M) {
	q.Client.Del("test", "test/proc")
	os.Exit(m.Run())
}

func TestPubSub(t *testing.T) {
	m := new(HandlerMock)

	q.Subscribe("test", m.handleFunc)
	m.AssertNotCalled(t, "handleFunc")

	m.On("handleFunc", []byte("message")).Return(nil)
	q.Publish("test", []byte("message"))
	time.Sleep(100 * time.Millisecond)
	m.AssertCalled(t, "handleFunc", []byte("message"))

	assert.Equal(t, int64(0), q.Client.LLen("test").Val())
	assert.Equal(t, int64(0), q.Client.LLen("test/proc").Val())
}

func TestPubSubFailure(t *testing.T) {
	m := new(HandlerMock)

	q.Subscribe("test", m.handleFunc)
	m.AssertNotCalled(t, "handleFunc")

	m.On("handleFunc", []byte("message")).Return(errors.New(""))
	q.Publish("test", []byte("message"))
	time.Sleep(100 * time.Millisecond)
	m.AssertCalled(t, "handleFunc", []byte("message"))

	assert.Equal(t, int64(0), q.Client.LLen("test").Val())
	assert.Equal(t, int64(1), q.Client.LLen("test/proc").Val())
}
