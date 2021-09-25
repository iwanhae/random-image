package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/iwanhae/random-image/pkg/server"
	"github.com/iwanhae/random-image/pkg/store"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/viper"
)

const (
	EnvPrefix = "RI"

	EnvS3Endpoint  = "S3_ENDPOINT"
	EnvS3AccessKey = "S3_ACCESS_KEY"
	EnvS3SecretKey = "S3_SECRET_KEY"
)

func GetMinioCliet(v *viper.Viper) minio.Client {
	endpoint := v.GetString(EnvS3Endpoint)
	accessKeyID := v.GetString(EnvS3AccessKey)
	secretAccessKey := v.GetString(EnvS3SecretKey)
	useSSL := false
	fmt.Println(endpoint, accessKeyID, secretAccessKey)
	mc, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatal(err)
	}
	return *mc
}

func main() {
	// config
	v := viper.New()
	v.SetEnvPrefix(EnvPrefix)
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		log.Println("Load from ENV")
		v.AutomaticEnv()
	} else {
		log.Println("Load from dotenv")
		v.SetConfigType("dotenv")
		v.SetConfigFile(".env")
		err := v.ReadInConfig()
		if err != nil {
			log.Fatalf("fail to read dotenv file:%s", err.Error())
		}
	}
	log.Println(v.AllSettings())

	db := store.NewDatabse()

	// minio
	ctx := context.Background()
	log.Println("Get Minio Client")
	mc := GetMinioCliet(v)

	go func() {
		ch := mc.ListObjects(ctx, "images", minio.ListObjectsOptions{Recursive: true})
		//
		tmp := []minio.ObjectInfo{}
		for v := range ch {
			tmp = append(tmp, v)
			if len(tmp)%1000 == 0 {
				db.PutBatchObjectMeta(ctx, tmp)
				fmt.Println(db.Stats(ctx))
				tmp = []minio.ObjectInfo{}
			}
		}
		db.PutBatchObjectMeta(ctx, tmp)
		//
		log.Println("done")
		log.Println(db.Stats(ctx))
	}()

	// echo
	e := server.NewServer(mc, db)
	e.Logger.Fatal(e.Start(":8080"))
}
