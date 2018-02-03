package pool

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/aquiladev/monday/storage"
	azurestorage "github.com/azure/azure-sdk-for-go/storage"
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
	log.Tracef("BLOBS %d", len(blobs))

	// get min length
	length := len(blobs)
	if length > q.numOfMessages {
		length = q.numOfMessages
	}

	messages := make([]*storage.Message, 0)
	ch := make(chan *storage.Message)
	defer close(ch)

	for i := 0; i < length; i++ {
		log.Tracef("BLOB %s", blobs[i].Name)
		go func(blob *azurestorage.Blob, msgCh chan *storage.Message) {
			content, err := q.blobRepo.Read(blob)
			if err != nil {
				log.Debug(err)
				msgCh <- nil
				return
			}

			msg := storage.Message{}
			err = json.Unmarshal(content, &msg)
			if err != nil {
				log.Debug(err)
				msgCh <- nil
				return
			}

			if err := q.blobRepo.Delete(blob.Name); err != nil {
				log.Debug(err)
				msgCh <- nil
				return
			}

			msgCh <- &msg
		}(&blobs[i], ch)
	}

	for i := 0; i < length; i++ {
		if msg := <-ch; msg != nil {
			messages = append(messages, msg)
		}
	}

	return messages, nil
}
