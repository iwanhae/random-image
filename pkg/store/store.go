package store

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	mime "github.com/cubewise-code/go-mime"
	"github.com/iwanhae/random-image/pkg/store/meta"
	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
)

type ObjectStore struct {
	Database    meta.Database
	MinioClient *minio.Client
	bucketName  string
}

func NewObjectStore(metaDB meta.Database, mc *minio.Client, bucketName string) (*ObjectStore, error) {
	if metaDB == nil {
		return nil, fmt.Errorf("metaDB should not be nil")
	}
	if mc == nil {
		return nil, fmt.Errorf("minio Client should not be nil")
	}
	return &ObjectStore{
		Database:    metaDB,
		MinioClient: mc,
		bucketName:  bucketName,
	}, nil
}

func (o *ObjectStore) GetReader(ctx context.Context, uuid string) (*meta.ObjectMeta, io.ReadSeekCloser, error) {
	obj, err := o.Database.Get(ctx, uuid)
	if err != nil {
		return nil, nil, err
	}
	s3obj, err := o.MinioClient.GetObject(ctx, o.bucketName, obj.S3Key, minio.GetObjectOptions{})
	if err != nil {
		return nil, nil, err
	}
	return obj, s3obj, nil
}

func (o *ObjectStore) PutObject(ctx context.Context, m meta.ObjectMeta, reader io.Reader) (string, error) {
	ext := path.Ext(m.Path)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		return "", fmt.Errorf("%q dose not contain a valid extention", m.Path)
	}
	m.S3Key = ""
	uuid, err := o.Database.Create(ctx, m)
	if err != nil {
		return "", err
	}

	f, err := ioutil.TempFile("", "*")
	if err != nil {
		return "", errors.Wrap(err, "fail to create tmp file")
	}
	defer func() {
		os.Remove(f.Name())
	}()

	_, err = io.Copy(f, reader)
	if err != nil {
		return "", errors.Wrap(err, "fail to wrtie tmp file")
	}

	s3Key := fmt.Sprintf("%s%s", uuid, ext)
	info, err := o.MinioClient.FPutObject(ctx, o.bucketName, s3Key, f.Name(), minio.PutObjectOptions{ContentType: mimeType})
	if err != nil {
		return "", errors.Wrap(err, "fail to upload file to s3 storage")
	}

	err = o.Database.Update(ctx, uuid, &meta.ObjectMeta{S3Key: info.Key})
	if err != nil {
		return "", errors.Wrap(err, "fail to upload file to s3 storage")
	}

	return uuid, nil
}
