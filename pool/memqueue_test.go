package pool

import (
	"testing"

	"github.com/aquiladev/monday/storage"
	"github.com/btcsuite/btclog"
	"github.com/stretchr/testify/assert"
)

type logWriter struct{}

func (logWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func TestOverCapacity(t *testing.T) {
	msgPool := MemQueue{capacity: 0}
	err := msgPool.Put(&storage.Message{})

	assert.NotEmpty(t, err)
}

func TestPutPop(t *testing.T) {
	// setup
	backendLog := btclog.NewBackend(logWriter{})
	UseLogger(backendLog.Logger("TEST"))

	msgPool := NewMemQueue(10)

	// put
	msgPool.Put(&storage.Message{})
	msgPool.Put(&storage.Message{})

	assert.Equal(t, 2, msgPool.Count())

	// pop
	res, _ := msgPool.Pop(2)
	if msgPool.Count() != 0 && len(res) == 2 {
		t.Error("Wrong amount")
	}
}
