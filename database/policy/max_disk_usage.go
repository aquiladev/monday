package policy

import (
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/disk"
)

type maxDiskUsage struct {
	path     string
	maxUsage float64
}

var _ StoragePolicy = (*maxDiskUsage)(nil)

func (m *maxDiskUsage) IsAccept() (bool, error) {
	usage, err := disk.Usage(m.path)
	if err != nil {
		return false, err
	}

	if usage.UsedPercent >= m.maxUsage {
		return false, errors.New("MaxDiskUsage Policy: reached max disk usage")
	}

	return true, nil
}

func NewMaxDiskUsage(path string, maxUsage float64) StoragePolicy {
	return &maxDiskUsage{
		path:     path,
		maxUsage: maxUsage,
	}
}
