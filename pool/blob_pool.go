package pool

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/aquiladev/monday/storage"
)

type BlobPool struct {
	blobRepo      *storage.AzureBlobRepository
	numOfMessages int
}

var _ Pool = (*BlobPool)(nil)

func NewBlobPool(blobRepo *storage.AzureBlobRepository, numOfMessages int) *BlobPool {
	return &BlobPool{
		blobRepo:      blobRepo,
		numOfMessages: numOfMessages,
	}
}

func (q *BlobPool) Put(msg *storage.Message) error {
	blobName := strconv.FormatInt(time.Now().Unix(), 10)
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return q.blobRepo.Insert(blobName, data)
}

func (q *BlobPool) Pop() ([]storage.Message, error) {
	blobs, err := q.blobRepo.Get()
	if err != nil {
		return nil, err
	}

	// get min length
	length := len(blobs)
	if length > q.numOfMessages {
		length = q.numOfMessages
	}

	messages := make([]storage.Message, length)
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

		messages[i] = msg
	}

	for i := 0; i < length; i++ {
		if err := q.blobRepo.Delete(blobs[i].Name); err != nil {
			return nil, err
		}
	}

	return messages, nil
}
