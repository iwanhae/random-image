package main

import (
	"archive/zip"
	"context"
	"encoding/base64"
	"fmt"
	"path"

	"github.com/iwanhae/random-image/pkg/config"
	"github.com/iwanhae/random-image/pkg/storage"
	"github.com/minio/minio-go/v7"
)

func main() {
	c := config.Init()

	mc, err := storage.GetS3Client(c)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	sem := make(chan int, 100)
	for val := range mc.ListObjects(ctx, c.S3BucketName, minio.ListObjectsOptions{Prefix: "zip", Recursive: true}) {
		if path.Ext(val.Key) == ".zip" {
			v := val
			sem <- 1
			go func() {
				group := base64.StdEncoding.EncodeToString([]byte(v.Key))
				obj, err := mc.GetObject(ctx, c.S3BucketName, v.Key, minio.GetObjectOptions{})
				if err != nil {
					panic(err)
				}
				r, err := zip.NewReader(obj, v.Size)
				if err != nil {
					panic(err)
				}
				for _, f := range r.File {
					fmt.Println(f.Name)
					reader, err := f.Open()
					if err != nil {
						panic(err)
					}
					dir, file := path.Split(f.Name)
					dirB := base64.StdEncoding.EncodeToString([]byte(dir))
					mc.PutObject(ctx, c.S3BucketName, path.Join(
						group, dirB, file,
					), reader, int64(f.UncompressedSize64), minio.PutObjectOptions{})
				}
				<-sem
			}()
		}
	}

}
