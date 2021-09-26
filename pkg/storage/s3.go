package storage

import (
	"github.com/iwanhae/random-image/pkg/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"
)

func GetS3Client(c config.RandomImageConfig) (*minio.Client, error) {
	mc, err := minio.New(c.S3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.S3AccessKey, c.S3SecrretKey, ""),
		Secure: c.S3EnableTLS,
	})
	if err != nil {
		return nil, errors.Wrap(err, "fail to get minio client")
	}
	return mc, nil
}
