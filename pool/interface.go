package pool

import "github.com/aquiladev/monday/storage"

type Pool interface {
	Put(msg *storage.Message) error
	Pop() ([]storage.Message, error)
}