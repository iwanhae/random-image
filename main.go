package main

import (
	"context"
	"path"
	"strings"

	mime "github.com/cubewise-code/go-mime"
	"github.com/iwanhae/random-image/pkg/config"
	"github.com/iwanhae/random-image/pkg/server"
	"github.com/iwanhae/random-image/pkg/storage"
	"github.com/iwanhae/random-image/pkg/store"
	"github.com/minio/minio-go/v7"
	"github.com/rs/zerolog/log"
)

const (
	EnvPrefix = "RI"

	EnvS3Endpoint  = "S3_ENDPOINT"
	EnvS3AccessKey = "S3_ACCESS_KEY"
	EnvS3SecretKey = "S3_SECRET_KEY"
)

func main() {
	// config
	c := config.Init()
	log.Info().Interface("config", c).Msg("config loaded")

	db := store.NewDatabse()

	// minio
	ctx := context.Background()
	mc, err := storage.GetS3Client(c)
	if err != nil {
		panic(err)
	}

	go func() {
		ch := mc.ListObjects(ctx, c.S3BucketName, minio.ListObjectsOptions{Recursive: true})
		//
		tmp := []minio.ObjectInfo{}
		for v := range ch {
			if !strings.HasPrefix(mime.TypeByExtension(path.Ext(v.Key)), "image") {
				continue
			}
			tmp = append(tmp, v)
			if len(tmp)%1000 == 0 {
				db.PutBatchObjectMeta(ctx, tmp)
				tmp = []minio.ObjectInfo{}
			}
		}
		db.PutBatchObjectMeta(ctx, tmp)
		//
		if s, err := db.Stats(ctx); err == nil {
			log.Info().Interface("db_stats", s).Msg("done")
		} else {
			log.Fatal().Err(err)
		}
	}()

	// echo
	e := server.NewServer(c, mc, db)
	e.Logger.Fatal(e.Start(":8080"))
}
