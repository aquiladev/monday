package pool

import (
	"sync"

	"github.com/aquiladev/monday/storage"
	"github.com/pkg/errors"
)

type MemQueue struct {
	mtx      sync.Mutex
	capacity int
	pool     []*storage.Message
}

func (mp *MemQueue) Put(msg *storage.Message) error {
	if mp.isFull() {
		return errors.New("Memory queue is full")
	}

	mp.mtx.Lock()
	defer mp.mtx.Unlock()
	mp.pool = append(mp.pool, msg)

	log.Tracef("Put message into memory queue %+v", msg)
	return nil
}

func (mp *MemQueue) Pop(numOfMsg int) ([]*storage.Message, error) {
	if numOfMsg < 1 {
		err := errors.New("Number os messages cannot be less then one")
		return []*storage.Message{}, err
	}

	length := mp.Count()
	if length == 0 {
		err := errors.New("Number os messages cannot be less then one")
		return []*storage.Message{}, err
	}

	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	// min(numOfMsg, length)
	i := numOfMsg
	if numOfMsg > length {
		i = length
	}

	messages := mp.pool[:i]
	mp.pool = mp.pool[i:]

	log.Tracef("Pop messages from memory queue %+v", messages)
	return messages, nil
}

func (mp *MemQueue) Count() int {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()
	return len(mp.pool)
}

func (mp *MemQueue) isFull() bool {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()
	return len(mp.pool) == mp.capacity
}

func NewMemQueue(capacity int) *MemQueue {
	return &MemQueue{
		capacity: capacity,
		pool:     make([]*storage.Message, 0),
	}
}
