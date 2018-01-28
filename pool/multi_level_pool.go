package pool

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/aquiladev/monday/storage"
)

type MultiLevelPool struct {
	memQueue      *MemQueue
	blobRepo      *storage.AzureBlobRepository
	numOfMessages int
	keepLocal     bool
}

var _ Pool = (*MultiLevelPool)(nil)

func NewMultiLevelPool(
	memQueue *MemQueue,
	blobRepo *storage.AzureBlobRepository,
	numOfMessages int,
	keepLocal bool) *MultiLevelPool {

	return &MultiLevelPool{
		memQueue:      memQueue,
		blobRepo:      blobRepo,
		numOfMessages: numOfMessages,
		keepLocal:     keepLocal,
	}
}

func (q *MultiLevelPool) Put(msg *storage.Message) error {
	// 1. put to mem queue
	if q.keepLocal {
		if err := q.memQueue.Put(msg); err == nil {
			return nil
		}
	}

	// 2. put to blob
	blobName := strconv.FormatInt(time.Now().Unix(), 10)
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return q.blobRepo.Insert(blobName, data)
}

func (q *MultiLevelPool) Pop() ([]*storage.Message, error) {
	// 1. pop from mem queue
	if q.keepLocal && q.memQueue.Count() > 0 {
		return q.memQueue.Pop(q.numOfMessages)
	}

	// 2. pop from blob
	blobs, err := q.blobRepo.Get()
	if err != nil {
		return nil, err
	}

	// get min length
	length := len(blobs)
	if length > q.numOfMessages {
		length = q.numOfMessages
	}

	messages := make([]*storage.Message, length)
	for i := 0; i < length; i++ {
		content, err := q.blobRepo.Read(&blobs[i])
		if err != nil {
			return nil, err
		}

		msg := storage.Message{}
		err = json.Unmarshal(content, &msg)
		if err != nil {
			return nil, err
		}

		messages[i] = &msg
	}

	for i := 0; i < length; i++ {
		if err := q.blobRepo.Delete(blobs[i].Name); err != nil {
			return nil, err
		}
	}

	return messages, nil
}
