package policy

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

type maxDbSize struct {
	path    string
	maxSize int64
}

var _ StoragePolicy = (*maxDbSize)(nil)

func (m *maxDbSize) IsAccept() (bool, error) {
	var size int64
	err := filepath.Walk(m.path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	if err != nil {
		return false, err
	}

	if size >= m.maxSize {
		return false, errors.New("MaxDbSize Policy: reached max database size")
	}
	return true, nil
}

// NewMaxDbSize creates policy
//
// path - path to directory with db
// maxSize - max allowed size of db in bytes
func NewMaxDbSize(path string, maxSize int64) StoragePolicy {
	return &maxDbSize{
		path:    path,
		maxSize: maxSize,
	}
}
