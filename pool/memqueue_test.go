package pool

import (
	"testing"

	"github.com/aquiladev/monday/storage"
	"github.com/btcsuite/btclog"
)

type logWriter struct{}

func (logWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func TestOverCapacity(t *testing.T) {
	msgPool := MemQueue{
		capacity: 0,
	}
	err := msgPool.Put(storage.Message{})
	if err == nil {
		t.Error("Wrong capacity")
	}
}

func TestPutPop(t *testing.T) {
	// setup
	backendLog := btclog.NewBackend(logWriter{})
	UseLogger(backendLog.Logger("TEST"))

	msgPool := NewMemQueue(10)

	// put
	msgPool.Put(data.Message{})
	msgPool.Put(data.Message{})

	if msgPool.Count() != 2 {
		t.Error("Wrong amount")
	}

	// pop
	res, _ := msgPool.Pop(2)
	if msgPool.Count() != 0 && len(res) == 2 {
		t.Error("Wrong amount")
	}
}
