package database

import (
	"github.com/btcsuite/goleveldb/leveldb"
)

type KeyValue struct {
	K []byte
	V []byte
}

type DB interface {
	Put([]byte, []byte) error
	PutBatch([]*KeyValue) error
	Get([]byte) ([]byte, error)
	Close() error
}

type db struct {
	ldb *leveldb.DB
}

var _ DB = (*db)(nil)

func (s *db) Put(k, v []byte) error {
	return s.ldb.Put(k, v, nil)
}

func (s *db) PutBatch(pairs []*KeyValue) error {
	batch := new(leveldb.Batch)
	for _, p := range pairs {
		batch.Put(p.K, p.V)
	}

	return s.ldb.Write(batch, nil)
}

func (s *db) Get(k []byte) ([]byte, error) {
	return s.ldb.Get(k, nil)
}

func (s *db) Close() error {
	return s.ldb.Close()
}

func NewDB(path string) (*db, error) {
	ldb, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}

	return &db{ldb: ldb}, nil
}
