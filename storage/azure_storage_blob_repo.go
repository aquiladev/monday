package storage

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/azure/azure-sdk-for-go/storage"
)

type AzureBlobRepository struct {
	container *storage.Container
}

func NewAzureBlobRepository(accountName, accountKey, containerName string) (*AzureBlobRepository, error) {
	client, _ := storage.NewBasicClient(accountName, accountKey)
	client.HTTPClient.Transport = &http.Transport{DisableKeepAlives: true}

	container, err := ensureContainer(client, containerName)
	if err != nil {
		return nil, err
	}

	return &AzureBlobRepository{
		container: container,
	}, nil
}

func ensureContainer(client storage.Client, containerName string) (*storage.Container, error) {
	service := client.GetBlobService()
	container := service.GetContainerReference(containerName)

	exists, err := container.Exists()
	if err != nil {
		return nil, err
	}

	if !exists {
		return container, container.Create(nil)
	}
	return container, nil
}

func (t *AzureBlobRepository) Get() ([]storage.Blob, error) {
	params := storage.ListBlobsParameters{}
	res, err := t.container.ListBlobs(params)
	if err != nil {
		return nil, err
	}

	return res.Blobs, nil
}

func (t *AzureBlobRepository) Read(blob *storage.Blob) ([]byte, error) {
	readCloser, err := blob.Get(nil)
	if err != nil {
		return nil, err
	}
	defer readCloser.Close()

	return ioutil.ReadAll(readCloser)
}

func (t *AzureBlobRepository) Insert(blobName string, data []byte) error {
	blob := t.container.GetBlobReference(blobName)
	return blob.CreateBlockBlobFromReader(bytes.NewReader(data), nil)
}

func (t *AzureBlobRepository) Delete(blobName string) error {
	blob := t.container.GetBlobReference(blobName)
	return blob.Delete(nil)
}
