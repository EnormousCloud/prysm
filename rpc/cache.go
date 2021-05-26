package rpc

import (
	"os"

	"github.com/EnormousCloud/objstorage/pkg/s3object"
)

type IStorage interface {
	Has(key string) bool
	Get(key string) ([]byte, error)
	Set(key string, buf []byte) error
}

func NewStorage() (IStorage, error) {
	return s3object.FromEnv(os.Getenv("AWS_S3_BUCKET"), "cache/")
	// return localobject.NewLocalObject("/tmp/cache"), nil
}
