package policy

type StoragePolicy interface {
	IsAccept() (bool, error)
}
