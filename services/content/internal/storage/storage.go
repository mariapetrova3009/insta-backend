package storage

type PutResult struct {
	Key  string // media_path
	Size int64
}

type Storage interface {
	Put(name string, data []byte, mime string) (PutResult, error)
	Delete(key string) error
}
